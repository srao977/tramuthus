package alpaca

import (
	"testing"
	"time"
)

func TestObservationFromRawPreservesBU(t *testing.T) {
	receipt := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	for _, msgType := range []string{"b", "u"} {
		raw := []byte(`{"T":"` + msgType + `","S":"AAPL","o":1,"h":2,"l":1,"c":1.5,"v":10,"t":"2026-09-10T12:00:00Z","n":3}`)
		obs, err := observationFromRaw(raw, "ALPACA_IEX", receipt)
		if err != nil {
			t.Fatal(err)
		}
		if obs.AlpacaMessageType != msgType {
			t.Fatalf("type=%s", obs.AlpacaMessageType)
		}
		if obs.Symbol != "AAPL" || obs.Volume != 10 || obs.Interval != "1Min" {
			t.Fatalf("%+v", obs)
		}
		if obs.PayloadHash == "" {
			t.Fatal("missing payload hash")
		}
	}
}

func TestMalformedRejected(t *testing.T) {
	receipt := time.Now().UTC()
	cases := []string{
		`{"T":"b","S":"","o":1,"h":2,"l":1,"c":1,"v":1,"t":"2026-09-10T12:00:00Z"}`,
		`{"T":"b","S":"AAPL","o":1,"h":2,"l":1,"c":1,"v":1,"t":""}`,
		`{"T":"b","S":"AAPL","o":0,"h":2,"l":1,"c":1,"v":1,"t":"2026-09-10T12:00:00Z"}`,
		`{"T":"b","S":"AAPL","o":3,"h":2,"l":1,"c":1,"v":1,"t":"2026-09-10T12:00:00Z"}`,
	}
	for _, raw := range cases {
		if _, err := observationFromRaw([]byte(raw), "ALPACA_IEX", receipt); err == nil {
			t.Fatalf("expected reject for %s", raw)
		}
	}
}

func TestDecodeMessageArray(t *testing.T) {
	msgs, err := decodeMessageArray([]byte(`[{"T":"success","msg":"authenticated"}]`))
	if err != nil || len(msgs) != 1 {
		t.Fatalf("%v %v", msgs, err)
	}
	c, err := decodeControl(msgs[0])
	if err != nil || c.Message != "authenticated" {
		t.Fatalf("%+v %v", c, err)
	}
}
