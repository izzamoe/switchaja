package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"switchiot/internal/db"
	"switchiot/internal/iot"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type API struct {
	DB     *sql.DB
	Sender iot.CommandSender
	Hub    *iot.Hub
	// dynamic mqtt
	Mqtt        *iot.MQTTSender
	mqttRetries int
}

func New(database *sql.DB, sender iot.CommandSender, hub *iot.Hub) *API {
	return &API{DB: database, Sender: sender, Hub: hub}
}

func (a *API) Register(app *fiber.App) {
	// -------- Public (no auth) --------
	app.Post("/login", a.login)
	// me & logout need a valid cookie but keep simple (they self-handle 401)
	app.Post("/logout", a.logout)
	app.Get("/me", a.me)

	// -------- Grouped Protected Routes --------
	userGroup := app.Group("/api/", a.authRequired("user"))   // any logged in user (role user/admin)
	adminGroup := app.Group("/api/", a.authRequired("admin")) // admin only

	// user capabilities (includes admin)
	userGroup.Post("start", a.start)
	userGroup.Post("extend", a.extend)
	userGroup.Post("stop", a.stop)
	userGroup.Get("status", a.status)
	userGroup.Get("transactions/:console_id", a.transactions)
	userGroup.Get("mqtt/status", a.mqttStatus)
	
	// reports endpoints
	userGroup.Get("reports/daily", a.dailyReport)
	userGroup.Get("reports/monthly", a.monthlyReport)
	userGroup.Get("reports/transactions", a.transactionReport)
	userGroup.Get("reports/export", a.exportTransactions)

	// admin only
	adminGroup.Get("users", a.listUsers)
	adminGroup.Post("users", a.createUser)
	adminGroup.Delete("users/:id", a.deleteUser)
	adminGroup.Post("price", a.updatePrice)
	adminGroup.Post("mqtt/config", a.mqttConfig)

	// Legacy routes without /api prefix for backward compatibility
	app.Post("/start", a.authRequired("user"), a.start)
	app.Post("/extend", a.authRequired("user"), a.extend)
	app.Post("/stop", a.authRequired("user"), a.stop)
	app.Get("/status", a.authRequired("user"), a.status)
	app.Get("/transactions/:console_id", a.authRequired("user"), a.transactions)
	app.Get("/mqtt/status", a.authRequired("user"), a.mqttStatus)
	app.Get("/reports/daily", a.authRequired("user"), a.dailyReport)
	app.Get("/reports/monthly", a.authRequired("user"), a.monthlyReport)
	app.Get("/reports/transactions", a.authRequired("user"), a.transactionReport)
	app.Get("/reports/export", a.authRequired("user"), a.exportTransactions)
	app.Get("/users", a.authRequired("admin"), a.listUsers)
	app.Post("/users", a.authRequired("admin"), a.createUser)
	app.Delete("/users/:id", a.authRequired("admin"), a.deleteUser)
	app.Post("/price", a.authRequired("admin"), a.updatePrice)
	app.Post("/mqtt/config", a.authRequired("admin"), a.mqttConfig)
}

// session cookie name
const sessionCookie = "hehetoken"

type sessionData struct {
	UserID   int64  `json:"uid"`
	Username string `json:"username"`
	Role     string `json:"role"`
	IssuedAt int64  `json:"iat"`
}

// very small in-memory session store (map token->session)
var sessions = struct{ m map[string]sessionData }{m: map[string]sessionData{}}

// IsValidToken checks if token exists (for root route guard)
func IsValidToken(tok string) bool {
	if tok == "" {
		return false
	}
	_, ok := sessions.m[tok]
	return ok
}

// generate simple random token (not cryptographically strong but ok for local LAN usage)
func newToken(u string) string { return fmt.Sprintf("%x", time.Now().UnixNano()) + "-" + u }

// authRequired middleware; minRole can be "user" or "admin". Admin satisfies all.
func (a *API) authRequired(minRole string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tok := c.Cookies(sessionCookie)
		if tok == "" {
			return fiber.NewError(http.StatusUnauthorized, "unauthenticated")
		}
		sd, ok := sessions.m[tok]
		if !ok {
			return fiber.NewError(http.StatusUnauthorized, "invalid session")
		}
		// role check
		if minRole == "admin" && sd.Role != "admin" {
			return fiber.NewError(http.StatusForbidden, "admin only")
		}
		// attach user context
		c.Locals("user", sd)
		return c.Next()
	}
}

func (a *API) login(c *fiber.Ctx) error {
	var body struct{ Username, Password string }
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(400, err.Error())
	}
	u, hash, ok, err := db.GetUserByUsername(a.DB, strings.TrimSpace(body.Username))
	if err != nil || !ok {
		return fiber.NewError(401, "login gagal")
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(body.Password)) != nil {
		return fiber.NewError(401, "login gagal")
	}
	tok := newToken(u.Username)
	sessions.m[tok] = sessionData{UserID: u.ID, Username: u.Username, Role: u.Role, IssuedAt: time.Now().Unix()}
	c.Cookie(&fiber.Cookie{Name: sessionCookie, Value: tok, HTTPOnly: true, Secure: false, SameSite: "Lax", Path: "/", Expires: time.Now().Add(12 * time.Hour)})
	return c.JSON(fiber.Map{"status": "ok", "user": u})
}

func (a *API) logout(c *fiber.Ctx) error {
	tok := c.Cookies(sessionCookie)
	if tok != "" {
		delete(sessions.m, tok)
	}
	c.Cookie(&fiber.Cookie{Name: sessionCookie, Value: "", Expires: time.Now().Add(-1 * time.Hour), Path: "/"})
	return c.JSON(fiber.Map{"status": "ok"})
}

func (a *API) me(c *fiber.Ctx) error {
	tok := c.Cookies(sessionCookie)
	sd, ok := sessions.m[tok]
	if !ok {
		return fiber.NewError(401, "unauthenticated")
	}
	return c.JSON(sd)
}

// user management handlers
func (a *API) listUsers(c *fiber.Ctx) error {
	users, err := db.ListUsers(a.DB)
	if err != nil {
		return fiber.NewError(500, err.Error())
	}
	return c.JSON(users)
}

func (a *API) createUser(c *fiber.Ctx) error {
	var body struct{ Username, Password, Role string }
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(400, err.Error())
	}
	body.Username = strings.TrimSpace(body.Username)
	if body.Username == "" || body.Password == "" {
		return fiber.NewError(400, "username/password required")
	}
	if body.Role == "" {
		body.Role = "user"
	}
	h, _ := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	id, err := db.CreateUser(a.DB, body.Username, string(h), body.Role)
	if err != nil {
		return fiber.NewError(400, err.Error())
	}
	return c.JSON(fiber.Map{"id": id})
}

func (a *API) deleteUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(400, "invalid id")
	}
	if err := db.DeleteUser(a.DB, int64(id)); err != nil {
		return fiber.NewError(400, err.Error())
	}
	return c.JSON(fiber.Map{"status": "deleted"})
}

func (a *API) start(c *fiber.Ctx) error {
	return a.withBroadcast(c, func() error {
		var body struct {
			ConsoleID   int64 `json:"console_id"`
			DurationMin int   `json:"duration_minutes"`
		}
		if err := c.BodyParser(&body); err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		if err := db.StartRental(a.DB, body.ConsoleID, body.DurationMin); err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		_ = a.Sender.Send(body.ConsoleID, "ON")
		return c.JSON(fiber.Map{"status": "ok"})
	})
}

func (a *API) extend(c *fiber.Ctx) error {
	return a.withBroadcast(c, func() error {
		var body struct {
			ConsoleID  int64 `json:"console_id"`
			AddMinutes int   `json:"add_minutes"`
		}
		if err := c.BodyParser(&body); err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		if err := db.ExtendRental(a.DB, body.ConsoleID, body.AddMinutes); err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		return c.JSON(fiber.Map{"status": "ok"})
	})
}

func (a *API) stop(c *fiber.Ctx) error {
	return a.withBroadcast(c, func() error {
		var body struct {
			ConsoleID int64 `json:"console_id"`
		}
		if err := c.BodyParser(&body); err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		if err := db.StopRental(a.DB, body.ConsoleID); err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		_ = a.Sender.Send(body.ConsoleID, "OFF")
		return c.JSON(fiber.Map{"status": "ok"})
	})
}

func (a *API) status(c *fiber.Ctx) error {
	consoles, err := db.GetConsoles(a.DB)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}
	now := time.Now()
	type item struct {
		db.Console
		RemainingSec    int             `json:"remaining_sec"`
		LastTransaction *db.Transaction `json:"last_transaction,omitempty"`
	}
	res := make([]item, 0, len(consoles))
	for _, cs := range consoles {
		left := 0
		if cs.Status == "RUNNING" && cs.EndTime.After(now) {
			left = int(cs.EndTime.Sub(now).Seconds())
		}
		var lt *db.Transaction
		if tr, ok, _ := db.LastTransaction(a.DB, cs.ID); ok {
			lt = &tr
		}
		res = append(res, item{Console: cs, RemainingSec: left, LastTransaction: lt})
	}
	return c.JSON(res)
}

func (a *API) transactions(c *fiber.Ctx) error {
	id, err := c.ParamsInt("console_id")
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "invalid id")
	}
	rows, err := a.DB.Query(`SELECT id, console_id, start_time, end_time, duration_minutes, total_price, COALESCE(price_per_hour_snapshot,0) FROM transactions WHERE console_id=? ORDER BY id DESC LIMIT 50`, id)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()
	var list []db.Transaction
	for rows.Next() {
		var t db.Transaction
		if err := rows.Scan(&t.ID, &t.ConsoleID, &t.StartTime, &t.EndTime, &t.DurationMin, &t.TotalPrice, &t.PricePerHourSnapshot); err != nil {
			return fiber.NewError(http.StatusInternalServerError, err.Error())
		}
		list = append(list, t)
	}
	return c.JSON(list)
}

// dailyReport returns daily summary of total hours and revenue
func (a *API) dailyReport(c *fiber.Ctx) error {
	dateParam := c.Query("date") // Format: YYYY-MM-DD
	if dateParam == "" {
		dateParam = time.Now().Format("2006-01-02")
	}
	
	// Parse the date
	date, err := time.Parse("2006-01-02", dateParam)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "invalid date format, use YYYY-MM-DD")
	}
	
	// Get start and end of the day
	startOfDay := date.Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24 * time.Hour)
	
	// Query for transactions in this day
	rows, err := a.DB.Query(`
		SELECT COALESCE(SUM(duration_minutes), 0) as total_minutes, 
		       COALESCE(SUM(total_price), 0) as total_revenue,
		       COUNT(*) as total_transactions
		FROM transactions 
		WHERE start_time >= ? AND start_time < ?`, 
		startOfDay, endOfDay)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()
	
	var totalMinutes, totalRevenue, totalTransactions int
	if rows.Next() {
		if err := rows.Scan(&totalMinutes, &totalRevenue, &totalTransactions); err != nil {
			return fiber.NewError(http.StatusInternalServerError, err.Error())
		}
	}
	
	// Convert minutes to hours
	totalHours := float64(totalMinutes) / 60.0
	
	return c.JSON(fiber.Map{
		"date": dateParam,
		"total_hours": totalHours,
		"total_revenue": totalRevenue,
		"total_transactions": totalTransactions,
	})
}

// monthlyReport returns monthly summary with revenue recap and hours per console
func (a *API) monthlyReport(c *fiber.Ctx) error {
	monthParam := c.Query("month") // Format: YYYY-MM
	if monthParam == "" {
		monthParam = time.Now().Format("2006-01")
	}
	
	// Parse the month
	month, err := time.Parse("2006-01", monthParam)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "invalid month format, use YYYY-MM")
	}
	
	// Get start and end of the month
	startOfMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0)
	
	// Query for overall monthly summary
	row := a.DB.QueryRow(`
		SELECT COALESCE(SUM(duration_minutes), 0) as total_minutes, 
		       COALESCE(SUM(total_price), 0) as total_revenue,
		       COUNT(*) as total_transactions
		FROM transactions 
		WHERE start_time >= ? AND start_time < ?`, 
		startOfMonth, endOfMonth)
	
	var totalMinutes, totalRevenue, totalTransactions int
	if err := row.Scan(&totalMinutes, &totalRevenue, &totalTransactions); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}
	
	// Query for per-console breakdown
	rows, err := a.DB.Query(`
		SELECT c.id, c.name, 
		       COALESCE(SUM(t.duration_minutes), 0) as total_minutes,
		       COALESCE(SUM(t.total_price), 0) as total_revenue,
		       COUNT(t.id) as total_transactions
		FROM consoles c
		LEFT JOIN transactions t ON c.id = t.console_id 
		    AND t.start_time >= ? AND t.start_time < ?
		GROUP BY c.id, c.name
		ORDER BY c.id`, 
		startOfMonth, endOfMonth)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()
	
	type ConsoleStats struct {
		ID               int64   `json:"console_id"`
		Name             string  `json:"console_name"`
		TotalHours       float64 `json:"total_hours"`
		TotalRevenue     int     `json:"total_revenue"`
		TotalTransactions int    `json:"total_transactions"`
	}
	
	var consoleStats []ConsoleStats
	for rows.Next() {
		var stat ConsoleStats
		var totalMinutes int
		if err := rows.Scan(&stat.ID, &stat.Name, &totalMinutes, &stat.TotalRevenue, &stat.TotalTransactions); err != nil {
			return fiber.NewError(http.StatusInternalServerError, err.Error())
		}
		stat.TotalHours = float64(totalMinutes) / 60.0
		consoleStats = append(consoleStats, stat)
	}
	
	totalHours := float64(totalMinutes) / 60.0
	
	return c.JSON(fiber.Map{
		"month": monthParam,
		"summary": fiber.Map{
			"total_hours": totalHours,
			"total_revenue": totalRevenue,
			"total_transactions": totalTransactions,
		},
		"console_breakdown": consoleStats,
	})
}

// transactionReport returns filtered transactions based on search criteria
func (a *API) transactionReport(c *fiber.Ctx) error {
	// Parse query parameters
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	consoleIDParam := c.Query("console_id")
	minAmount := c.Query("min_amount")
	maxAmount := c.Query("max_amount")
	
	// Build the query dynamically
	query := `SELECT t.id, t.console_id, c.name as console_name, t.start_time, t.end_time, 
	                 t.duration_minutes, t.total_price, t.price_per_hour_snapshot 
	          FROM transactions t 
	          JOIN consoles c ON t.console_id = c.id 
	          WHERE 1=1`
	
	var args []interface{}
	
	// Add date filters
	if dateFrom != "" {
		if _, err := time.Parse("2006-01-02", dateFrom); err != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid date_from format, use YYYY-MM-DD")
		}
		query += " AND DATE(t.start_time) >= ?"
		args = append(args, dateFrom)
	}
	
	if dateTo != "" {
		if _, err := time.Parse("2006-01-02", dateTo); err != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid date_to format, use YYYY-MM-DD")
		}
		query += " AND DATE(t.start_time) <= ?"
		args = append(args, dateTo)
	}
	
	// Add console filter
	if consoleIDParam != "" {
		query += " AND t.console_id = ?"
		args = append(args, consoleIDParam)
	}
	
	// Add amount filters
	if minAmount != "" {
		query += " AND t.total_price >= ?"
		args = append(args, minAmount)
	}
	
	if maxAmount != "" {
		query += " AND t.total_price <= ?"
		args = append(args, maxAmount)
	}
	
	query += " ORDER BY t.start_time DESC LIMIT 100"
	
	rows, err := a.DB.Query(query, args...)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()
	
	type TransactionDetail struct {
		ID                   int64     `json:"id"`
		ConsoleID            int64     `json:"console_id"`
		ConsoleName          string    `json:"console_name"`
		StartTime            time.Time `json:"start_time"`
		EndTime              time.Time `json:"end_time"`
		DurationMin          int       `json:"duration_minutes"`
		TotalPrice           int       `json:"total_price"`
		PricePerHourSnapshot int       `json:"price_per_hour"`
	}
	
	var transactions []TransactionDetail
	for rows.Next() {
		var t TransactionDetail
		if err := rows.Scan(&t.ID, &t.ConsoleID, &t.ConsoleName, &t.StartTime, &t.EndTime, 
			&t.DurationMin, &t.TotalPrice, &t.PricePerHourSnapshot); err != nil {
			return fiber.NewError(http.StatusInternalServerError, err.Error())
		}
		transactions = append(transactions, t)
	}
	
	return c.JSON(transactions)
}

// exportTransactions exports transactions to CSV format
func (a *API) exportTransactions(c *fiber.Ctx) error {
	// Use the same filtering logic as transactionReport
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	consoleIDParam := c.Query("console_id")
	minAmount := c.Query("min_amount")
	maxAmount := c.Query("max_amount")
	
	query := `SELECT t.id, t.console_id, c.name as console_name, t.start_time, t.end_time, 
	                 t.duration_minutes, t.total_price, t.price_per_hour_snapshot 
	          FROM transactions t 
	          JOIN consoles c ON t.console_id = c.id 
	          WHERE 1=1`
	
	var args []interface{}
	
	// Add filters (same logic as transactionReport)
	if dateFrom != "" {
		if _, err := time.Parse("2006-01-02", dateFrom); err != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid date_from format, use YYYY-MM-DD")
		}
		query += " AND DATE(t.start_time) >= ?"
		args = append(args, dateFrom)
	}
	
	if dateTo != "" {
		if _, err := time.Parse("2006-01-02", dateTo); err != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid date_to format, use YYYY-MM-DD")
		}
		query += " AND DATE(t.start_time) <= ?"
		args = append(args, dateTo)
	}
	
	if consoleIDParam != "" {
		query += " AND t.console_id = ?"
		args = append(args, consoleIDParam)
	}
	
	if minAmount != "" {
		query += " AND t.total_price >= ?"
		args = append(args, minAmount)
	}
	
	if maxAmount != "" {
		query += " AND t.total_price <= ?"
		args = append(args, maxAmount)
	}
	
	query += " ORDER BY t.start_time DESC"
	
	rows, err := a.DB.Query(query, args...)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()
	
	// Set CSV headers
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename=transactions.csv")
	
	// Write CSV header
	csvData := "ID,Console ID,Console Name,Start Time,End Time,Duration (Minutes),Total Price,Price Per Hour\n"
	
	// Write CSV data
	for rows.Next() {
		var id, consoleID, durationMin, totalPrice, pricePerHour int64
		var consoleName string
		var startTime, endTime time.Time
		
		if err := rows.Scan(&id, &consoleID, &consoleName, &startTime, &endTime, 
			&durationMin, &totalPrice, &pricePerHour); err != nil {
			return fiber.NewError(http.StatusInternalServerError, err.Error())
		}
		
		csvData += fmt.Sprintf("%d,%d,%s,%s,%s,%d,%d,%d\n",
			id, consoleID, consoleName,
			startTime.Format("2006-01-02 15:04:05"),
			endTime.Format("2006-01-02 15:04:05"),
			durationMin, totalPrice, pricePerHour)
	}
	
	return c.SendString(csvData)
}

func (a *API) updatePrice(c *fiber.Ctx) error {
	return a.withBroadcast(c, func() error {
		var body struct {
			ConsoleID int64 `json:"console_id"`
			Price     int   `json:"price_per_hour"`
		}
		if err := c.BodyParser(&body); err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		if err := db.UpdatePrice(a.DB, body.ConsoleID, body.Price); err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		return c.JSON(fiber.Map{"status": "ok"})
	})
}

// helper to broadcast after mutating operations
func (a *API) withBroadcast(c *fiber.Ctx, fn func() error) error {
	err := fn()
	if err == nil {
		go a.BroadcastStatus()
	}
	return err
}

func (a *API) BroadcastStatus() {
	if a.Hub == nil {
		return
	}
	a.Hub.Broadcast(a.StatusPayload())
}

// StatusPayload builds the current consoles snapshot as websocket message bytes.
func (a *API) StatusPayload() []byte {
	consoles, err := db.GetConsoles(a.DB)
	if err != nil {
		return []byte("{}")
	}
	now := time.Now()
	type item struct {
		db.Console
		RemainingSec    int             `json:"remaining_sec"`
		LastTransaction *db.Transaction `json:"last_transaction,omitempty"`
	}
	res := make([]item, 0, len(consoles))
	for _, cs := range consoles {
		left := 0
		if cs.Status == "RUNNING" && cs.EndTime.After(now) {
			left = int(cs.EndTime.Sub(now).Seconds())
		}
		var lt *db.Transaction
		if tr, ok, _ := db.LastTransaction(a.DB, cs.ID); ok {
			lt = &tr
		}
		res = append(res, item{Console: cs, RemainingSec: left, LastTransaction: lt})
	}
	b, _ := json.Marshal(fiber.Map{"type": "status", "data": res})
	return b
}

// mqttStatus returns JSON with connection info.
func (a *API) mqttStatus(c *fiber.Ctx) error {
	connected := false
	prefix := ""
	if a.Mqtt != nil {
		connected = a.Mqtt.IsConnected()
		prefix = a.Mqtt.Prefix()
	}
	return c.JSON(fiber.Map{"connected": connected, "prefix": prefix, "retries": a.mqttRetries})
}

// mqttConfig accepts broker/prefix/username/password; empty broker means disconnect.
func (a *API) mqttConfig(c *fiber.Ctx) error {
	type bodyT struct {
		Broker   string `json:"broker"`
		Prefix   string `json:"prefix"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var b bodyT
	if err := c.BodyParser(&b); err != nil {
		return fiber.NewError(400, err.Error())
	}
	if b.Broker == "" { // disconnect
		if a.Mqtt != nil {
			a.Mqtt.Close()
			a.Mqtt = nil
		}
		a.mqttRetries = 0
		_ = db.SaveMQTTConfig(a.DB, db.MQTTConfig{})
		a.broadcastMQTT()
		return c.JSON(fiber.Map{"status": "disconnected"})
	}
	cfg := db.MQTTConfig{Broker: b.Broker, Prefix: b.Prefix, Username: b.Username, Password: b.Password}
	_ = db.SaveMQTTConfig(a.DB, cfg)
	// attempt connect with up to 3 retries sequentially
	a.mqttRetries = 0
	var lastErr error
	for a.mqttRetries < 3 {
		opts := iot.MQTTSenderOptions{Prefix: b.Prefix, Username: b.Username, Password: b.Password, QOS: 1, CleanSession: true, StatusCallback: func(id int64, payload string) {}}
		mqttSender, err := iot.NewMQTTSender(b.Broker, opts)
		if err == nil {
			if a.Mqtt != nil {
				a.Mqtt.Close()
			}
			a.Mqtt = mqttSender
			a.Sender = iot.NewIdempotentSender(mqttSender)
			a.broadcastMQTT()
			return c.JSON(fiber.Map{"status": "connected", "retries": a.mqttRetries + 1})
		}
		lastErr = err
		a.mqttRetries++
		time.Sleep(1 * time.Second)
	}
	a.broadcastMQTT()
	return fiber.NewError(400, "mqtt connect failed after retries: "+lastErr.Error())
}

func (a *API) broadcastMQTT() {
	if a.Hub == nil {
		return
	}
	connected := false
	prefix := ""
	retries := a.mqttRetries
	if a.Mqtt != nil {
		connected = a.Mqtt.IsConnected()
		prefix = a.Mqtt.Prefix()
	}
	b, _ := json.Marshal(fiber.Map{"type": "mqtt", "connected": connected, "prefix": prefix, "retries": retries})
	a.Hub.Broadcast(b)
}
