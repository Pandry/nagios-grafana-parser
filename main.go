package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/url"
	"strconv"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
)

var (
	tstypes = []string{"last_update", "last_check", "last_state_change", "last_hard_state_change", "last_time_ok", "last_time_warning", "last_time_unknown", "last_time_critical", "last_notification", "next_notification", "next_check"}
	//jsonCols = []string{"host_name", "description", "plugin_output", "long_plugin_output", "perf_data", "max_attempts", "current_attempt", "status", "last_update", "has_been_checked", "should_be_scheduled", "last_check", "check_options", "check_type", "checks_enabled", "last_state_change", "last_hard_state_change", "last_hard_state", "last_time_ok", "last_time_warning", "last_time_unknown", "last_time_critical", "state_type", "last_notification", "next_notification", "next_check", "no_more_notifications", "notifications_enabled", "problem_has_been_acknowledged", "acknowledgement_type", "current_notification_number", "accept_passive_checks", "event_handler_enabled", "flap_detection_enabled", "is_flapping", "percent_state_change", "latency", "execution_time", "scheduled_downtime_depth", "process_performance_data", "obsess"}
)

func testConn(ctx *fasthttp.RequestCtx) {
	log.Println("/: Hit")
	fmt.Fprint(ctx, "OK")
}

func query(ctx *fasthttp.RequestCtx) {
	log.Println("/query: Hit")

	host, _ := url.Parse(viper.GetString("instances.nagiosinstance02.uri"))

	c := &fasthttp.HostClient{
		Addr: host.Hostname(),
	}

	req := &fasthttp.Request{}
	req.SetHost(host.Hostname())
	req.SetRequestURI(viper.GetString("instances.nagiosinstance02.uri") + `cgi-bin/statusjson.cgi?` + viper.GetString("querystring"))
	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(viper.GetString("instances.nagiosinstance02.username")+":"+viper.GetString("instances.nagiosinstance02.password"))))

	resp := &fasthttp.Response{}

	// Fetch google page via local proxy.
	err, statusCode, body := c.Do(req, resp), resp.StatusCode(), resp.Body()
	if err != nil {
		log.Fatalf("Error when loading nagios through local proxy: %s", err)
	}
	if statusCode != fasthttp.StatusOK {
		log.Fatalf("Unexpected status code: %d. Expecting %d", statusCode, fasthttp.StatusOK)
	}

	var p fastjson.Parser
	v, err := p.ParseBytes(body)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("foo=%s\n", v.GetStringBytes("str"))
	ctx.Response.Header.Add("Content-Type", "application/json")
	data := v.Get("data")
	columns := ""

	brk := false

	types := make([]fastjson.Type, 0)

	data.GetObject("servicelist").Visit(func(jj []byte, vv *fastjson.Value) {
		if brk {
			return
		}

		parent, _ := vv.Object()
		parent.Visit(func(jjj []byte, vvv *fastjson.Value) {
			if brk {
				return
			}
			brk = true

			obj, _ := vvv.Object()
			obj.Visit(func(key []byte, v *fastjson.Value) {
				//columns += `{"text":"` + string(key) + `","type":"string"},`
				keystring := string(key)
				types = append(types, v.Type())
				for _, v := range tstypes {
					if v == keystring {
						columns += `{"text":"` + keystring + `","type":"time"},`
						return
					}
				}
				columns += `{"text":"` + keystring + `","type":"` + getJSONTypeString(v.Type()) + `"},`

			})

		})
	})
	if len(columns) > 2 {
		columns = columns[:len(columns)-1]
	}

	//services := data.Get("servicelist")

	fmt.Fprint(ctx, `[{"columns":[`+columns+`]`+`,"rows":[`)

	first := true

	data.GetObject("servicelist").Visit(func(key []byte, vv *fastjson.Value) {
		//fmt.Fprint(ctx, v.String())
		obj, _ := vv.Object()
		obj.Visit(func(key []byte, row *fastjson.Value) {
			if !first {
				fmt.Fprint(ctx, `,`)
			}
			first = false
			fmt.Fprint(ctx, `[`)
			cols, _ := row.Object()
			i := 0

			cols.Visit(func(key []byte, col *fastjson.Value) {
				if getJSONTypeString(types[i]) == "string" {
					fmt.Fprint(ctx, col.String())
				} else {
					fmt.Fprint(ctx, col.String())
				}
				if i != len(types)-1 {
					fmt.Fprint(ctx, `,`)
				}
				i++
			})
			fmt.Fprint(ctx, `]`)
			//fmt.Fprint(ctx, v.String())
		})
	})

	fmt.Fprint(ctx, `],"type":"table"}]`)
}

/*
func Hello(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "hello, %s!\n", ctx.UserValue("name"))
}*/

func search(ctx *fasthttp.RequestCtx) {
	log.Println("/search: Hit")
	ctx.Response.Header.Add("Content-Type", "application/json")
	fmt.Fprint(ctx, `["host_name","description","plugin_output","long_plugin_output","perf_data","max_attempts","current_attempt","status","last_update","has_been_checked","should_be_scheduled","last_check","check_options","check_type","checks_enabled","last_state_change","last_hard_state_change","last_hard_state","last_time_ok","last_time_warning","last_time_unknown","last_time_critical","state_type","last_notification","next_notification","next_check","no_more_notifications","notifications_enabled","problem_has_been_acknowledged","acknowledgement_type","current_notification_number","accept_passive_checks","event_handler_enabled","flap_detection_enabled","is_flapping","percent_state_change","latency","execution_time","scheduled_downtime_depth","process_performance_data","obsess"]`)
}

func annotation(ctx *fasthttp.RequestCtx) {
	log.Println("/annotation: Hit")
}

func webApp(ctx *fasthttp.RequestCtx) {
	log.Println("/: Hit")
}

func main() {
	user := "admin"
	pass := "admin"

	pflag.Int("port", 8080, "Webserver listening port number")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(".")      // look for config in the working directory
	err := viper.ReadInConfig()   // Find and read the config file
	if err != nil {               // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error - cannot read config file: %s", err))
	}
	//viper.WriteConfig() // writes current config to predefined path set by 'viper.AddConfigPath()' and 'viper.SetConfigName'

	router := fasthttprouter.New()
	router.GET("/api", testConn)
	router.POST("/api/", testConn)
	router.GET("/api/search", search)
	router.POST("/api/search", search)
	router.GET("/api/query", query)
	router.POST("/api/query", query)
	router.GET("/api/annotations", annotation)
	router.GET("/js/", BasicAuth(fasthttp.FSHandler("./static/js/", 0), user, pass))
	router.GET("/", BasicAuth(fasthttp.FSHandler("./static", 0), user, pass))
	router.POST("/", BasicAuth(webApp, user, pass))
	router.NotFound = BasicAuth(fasthttp.FSHandler("./static", 0), user, pass)

	//router.GET("/hello/:name", Hello)

	log.Fatal(fasthttp.ListenAndServe(":"+strconv.Itoa(viper.GetInt("port")), router.Handler))
}

func getJSONTypeString(t fastjson.Type) string {
	switch t {
	//case typeRawString:
	case fastjson.TypeObject:
	case fastjson.TypeArray:
	case fastjson.TypeString:
		return "string"
	case fastjson.TypeNumber:
		return "number"
	case fastjson.TypeTrue:
	case fastjson.TypeFalse:
		return "string"
		//return "boolean"
	case fastjson.TypeNull:
		return "null"
	}
	return "string"
}
