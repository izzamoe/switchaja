package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"switchiot/internal/db"
	"switchiot/internal/iot"

	"github.com/gofiber/fiber/v2"
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
	app.Post("/start", a.start)
	app.Post("/extend", a.extend)
	app.Post("/stop", a.stop)
	app.Get("/status", a.status)
	app.Get("/transactions/:console_id", a.transactions)
	app.Post("/price", a.updatePrice)
	app.Get("/mqtt/status", a.mqttStatus)
	app.Post("/mqtt/config", a.mqttConfig)
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
