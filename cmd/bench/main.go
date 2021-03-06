package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"net"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/isucon/isucon9-qualify/bench/asset"
	"github.com/isucon/isucon9-qualify/bench/fails"
	"github.com/isucon/isucon9-qualify/bench/scenario"
	"github.com/isucon/isucon9-qualify/bench/server"
	"github.com/isucon/isucon9-qualify/bench/session"
)

type Output struct {
	Pass     bool     `json:"pass"`
	Score    int64    `json:"score"`
	Campaign int      `json:"campaign"`
	Language string   `json:"language"`
	Messages []string `json:"messages"`
}

type Config struct {
	TargetURLStr string
	TargetHost   string
	ShipmentURL  string
	PaymentURL   string
	PaymentPort  int
	ShipmentPort int

	AllowedIPs []net.IP
}

func init() {
	rand.Seed(time.Now().UnixNano())

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	flags := flag.NewFlagSet("isucon9q", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)

	conf := Config{}
	allowedIPStr := ""
	dataDir := ""
	staticDir := ""

	flags.StringVar(&conf.TargetURLStr, "target-url", "http://127.0.0.1:8000", "target url")
	flags.StringVar(&conf.TargetHost, "target-host", "isucon9.catatsuy.org", "target host")
	flags.StringVar(&conf.PaymentURL, "payment-url", "http://localhost:5555", "payment url")
	flags.StringVar(&conf.ShipmentURL, "shipment-url", "http://localhost:7000", "shipment url")
	flags.IntVar(&conf.PaymentPort, "payment-port", 5555, "payment service port")
	flags.IntVar(&conf.ShipmentPort, "shipment-port", 7000, "shipment service port")
	flags.StringVar(&dataDir, "data-dir", "initial-data", "data directory")
	flags.StringVar(&staticDir, "static-dir", "webapp/public/static", "static file directory")
	flags.StringVar(&allowedIPStr, "allowed-ips", "", "allowed ips (comma separated)")

	err := flags.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	if allowedIPStr != "" {
		for _, str := range strings.Split(allowedIPStr, ",") {
			aip := net.ParseIP(str)
			if aip == nil {
				log.Fatalf("allowed-ips: %s cannot be parsed", str)
			}
			conf.AllowedIPs = append(conf.AllowedIPs, aip)
		}
	}

	// ???????????????????????????
	sp, ss, err := server.RunServer(conf.PaymentPort, conf.ShipmentPort, dataDir, conf.AllowedIPs)
	if err != nil {
		log.Fatal(err)
	}

	scenario.SetShipment(ss)
	scenario.SetPayment(sp)

	err = session.SetShareTargetURLs(
		conf.TargetURLStr,
		conf.TargetHost,
		conf.PaymentURL,
		conf.ShipmentURL,
	)
	if err != nil {
		log.Fatal(err)
	}

	// ????????????????????????
	asset.Initialize(dataDir, staticDir)
	scenario.InitSessionPool()

	log.Print("=== initialize ===")
	// ????????????/initialize ????????????????????????????????????????????????????????????URL??????????????????DB?????????????????????????????????????????????
	campaign, language := scenario.Initialize(context.Background(), session.ShareTargetURLs.PaymentURL.String(), session.ShareTargetURLs.ShipmentURL.String())
	eMsgs := fails.ErrorsForCheck.GetMsgs()
	if len(eMsgs) > 0 {
		log.Print("cause error!")

		output := Output{
			Pass:     false,
			Score:    0,
			Campaign: campaign,
			Language: language,
			Messages: eMsgs,
		}
		json.NewEncoder(os.Stdout).Encode(output)

		return
	}

	log.Print("=== verify ===")
	// ????????????????????????????????????????????????????????????????????????
	// ????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????
	scenario.Verify(context.Background())
	eMsgs = fails.ErrorsForCheck.GetMsgs()
	if len(eMsgs) > 0 {
		log.Print("cause error!")

		output := Output{
			Pass:     false,
			Score:    0,
			Campaign: campaign,
			Language: language,
			Messages: eMsgs,
		}
		json.NewEncoder(os.Stdout).Encode(output)

		return
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(scenario.ExecutionSeconds*time.Second))
	defer cancel()

	log.Print("=== validation ===")

	// ?????????????????????????????????????????????
	// verify????????????????????????????????????????????????????????????????????????Validation????????????
	ss.SetDelay(800 * time.Millisecond)
	sp.SetDelay(800 * time.Millisecond)

	// ?????????????????????????????????check???load????????????2?????????????????????
	// check?????????????????????????????????????????????????????????????????????????????????
	// ????????????????????????????????????check???????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????
	// check???load????????????????????????????????????????????????????????????load?????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????
	// ?????????????????????????????????????????????????????????????????????????????????check???load??????????????????????????????
	scenario.Validation(ctx, campaign)

	// context.Canceled??????????????????????????????????????????????????????????????????
	eMsgs, cCnt, aCnt, tCnt := fails.ErrorsForCheck.Get()
	// critical error???1?????????????????????application error???10??????????????????
	if cCnt > 0 || aCnt >= 10 {
		log.Print("cause error!")

		output := Output{
			Pass:     false,
			Score:    0,
			Campaign: campaign,
			Language: language,
			Messages: uniqMsgs(eMsgs),
		}
		json.NewEncoder(os.Stdout).Encode(output)

		return
	}

	<-time.After(1 * time.Second)

	log.Print("=== final check ===")
	// ???????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????
	score := scenario.FinalCheck(context.Background())

	// application error?????????????????????
	fMsgs, _, faCnt, _ := fails.ErrorsForFinal.Get()
	msgs := append(uniqMsgs(eMsgs), fMsgs...)

	aCnt += faCnt

	// application error???10??????????????????
	if aCnt >= 10 {
		output := Output{
			Pass:     false,
			Score:    0,
			Campaign: campaign,
			Language: language,
			Messages: msgs,
		}
		json.NewEncoder(os.Stdout).Encode(output)

		return
	}

	// application error???1??????500?????????
	penalty := int64(500 * aCnt)

	if tCnt > 200 {
		// trivial error???200??????????????????100?????????5000?????????
		penalty += int64(5000 * (1 + (tCnt-200)/100))
	}

	log.Print(score, penalty)

	score -= penalty

	// 0?????????????????????
	if score <= 0 {
		output := Output{
			Pass:     false,
			Score:    0,
			Campaign: campaign,
			Language: language,
			Messages: msgs,
		}
		json.NewEncoder(os.Stdout).Encode(output)

		return
	}

	output := Output{
		Pass:     true,
		Score:    score,
		Campaign: campaign,
		Language: language,
		Messages: msgs,
	}
	json.NewEncoder(os.Stdout).Encode(output)
}

func uniqMsgs(allMsgs []string) []string {
	sort.Strings(allMsgs)
	msgs := make([]string, 0, len(allMsgs))

	tmp := ""

	// ?????????uniq??????
	for _, m := range allMsgs {
		if tmp != m {
			tmp = m
			msgs = append(msgs, m)
		}
	}

	return msgs
}
