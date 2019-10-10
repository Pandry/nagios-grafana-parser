package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/url"
	"strconv"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
)

const (
	nagiosAPIEndpoint = `cgi-bin/statusjson.cgi?`
)

//TODO:
//Convert to string the error? https://sourceforge.net/p/nagios/mailman/nagios-users/?limit=250&style=flat&viewmonth=200701&page=0

var (
	tstypes = []string{"last_update", "last_check", "last_state_change", "last_hard_state_change", "last_time_ok", "last_time_warning", "last_time_unknown", "last_time_critical", "last_notification", "next_notification", "next_check"}
	//jsonCols = []string{"host_name", "description", "plugin_output", "long_plugin_output", "perf_data", "max_attempts", "current_attempt", "status", "last_update", "has_been_checked", "should_be_scheduled", "last_check", "check_options", "check_type", "checks_enabled", "last_state_change", "last_hard_state_change", "last_hard_state", "last_time_ok", "last_time_warning", "last_time_unknown", "last_time_critical", "state_type", "last_notification", "next_notification", "next_check", "no_more_notifications", "notifications_enabled", "problem_has_been_acknowledged", "acknowledgement_type", "current_notification_number", "accept_passive_checks", "event_handler_enabled", "flap_detection_enabled", "is_flapping", "percent_state_change", "latency", "execution_time", "scheduled_downtime_depth", "process_performance_data", "obsess"}
	tableHeader       = []string{}
	tableHeaderString = ""
)

func testConn(ctx *fasthttp.RequestCtx) {
	log.Println("/: Hit")
	fmt.Fprint(ctx, "OK")
}

func query(ctx *fasthttp.RequestCtx) {
	log.Println("/query: Hit")
	instancesMap := viper.Get("instances").(map[string]interface{})
	instances := make([]string, 0, len(instancesMap))
	for k := range instancesMap {
		instances = append(instances, k)
	}

	for iter, instance := range instances {

		host, _ := url.Parse(viper.GetString("instances." + instance + ".uri"))

		c := &fasthttp.HostClient{
			Addr: host.Hostname(),
		}

		req := &fasthttp.Request{}
		req.SetHost(host.Hostname())
		//Check for custom query string against nagios
		settingPrepend := ""
		if viper.GetString("instances."+instance+".querystring") != "" {
			settingPrepend = "instances." + instance + "."
		}
		//Prepare the request
		req.SetRequestURI(viper.GetString("instances."+instance+".uri") + nagiosAPIEndpoint + viper.GetString(settingPrepend+"querystring"))

		if viper.GetBool("instances." + instance + ".authentication") {
			req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(viper.GetString("instances."+instance+".username")+":"+viper.GetString("instances."+instance+".password"))))
		}
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
		if iter == 0 {
			ctx.Response.Header.Add("Content-Type", "application/json")
			fmt.Fprint(ctx, `[{"columns":[`+tableHeaderString+`],"rows":[`)
		} else {
			fmt.Fprint(ctx, `,`)
		}
		data := v.Get("data")

		first := true

		data.GetObject("servicelist").Visit(func(key []byte, vv *fastjson.Value) {
			//fmt.Fprint(ctx, v.String())
			/*
				if !first && iter != 0 {
					fmt.Fprint(ctx, `,`)
				}*/
			obj, _ := vv.Object()
			obj.Visit(func(key []byte, row *fastjson.Value) {
				if row.GetBool("no_more_notifications") || (viper.GetBool("instances."+instance+".hideAcknowledged") || viper.GetBool("hideAcknowledged")) && row.GetBool("problem_has_been_acknowledged") {
					return
				}
				if !first {
					fmt.Fprint(ctx, `,`)
				} else {
					first = false
				}

				fmt.Fprint(ctx, `[`)

				for i, hv := range tableHeader {
					//o := row.GetObject(hv)

					//if viper.GetString("types."+hv) == "string" {
					tmpVal := row.Get(hv).String()
					if tmpVal == "" {
						tmpVal = `""`
					} else if maxLen := viper.GetInt("maxStringLength"); maxLen > 0 && len(tmpVal) > maxLen {
						tmpVal = tmpVal[:maxLen-1] + `"`
					}
					fmt.Fprint(ctx, tmpVal)
					/*} else {
						fmt.Fprint(ctx, row.Get(hv).String())
					}*/
					if i < len(tableHeader)-1 {
						fmt.Fprint(ctx, `,`)
					}

				}
				fmt.Fprint(ctx, `]`)
				/*
					if iter < len(instances)-1 {
						fmt.Fprint(ctx, `,`)
					}*/
			})
		})

	}
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

	viper.SetDefault("types.host_name", "string")
	viper.SetDefault("types.description", "string")
	viper.SetDefault("types.plugin_output", "string")
	viper.SetDefault("types.long_plugin_output", "string")
	viper.SetDefault("types.perf_data", "string")
	viper.SetDefault("types.max_attempts", "string")
	viper.SetDefault("types.current_attempt", "string")
	viper.SetDefault("types.status", "string")
	viper.SetDefault("types.last_update", "time")
	viper.SetDefault("types.has_been_checked", "string")
	viper.SetDefault("types.should_be_scheduled", "string")
	viper.SetDefault("types.last_check", "time")
	viper.SetDefault("types.check_options", "string")
	viper.SetDefault("types.check_type", "string")
	viper.SetDefault("types.checks_enabled", "string")
	viper.SetDefault("types.last_state_change", "time")
	viper.SetDefault("types.last_hard_state_change", "time")
	viper.SetDefault("types.last_hard_state", "string")
	viper.SetDefault("types.last_time_ok", "time")
	viper.SetDefault("types.last_time_warning", "time")
	viper.SetDefault("types.last_time_unknown", "time")
	viper.SetDefault("types.last_time_critical", "time")
	viper.SetDefault("types.state_type", "string")
	viper.SetDefault("types.last_notification", "time")
	viper.SetDefault("types.next_notification", "time")
	viper.SetDefault("types.next_check", "time")
	viper.SetDefault("types.no_more_notifications", "string")
	viper.SetDefault("types.notifications_enabled", "string")
	viper.SetDefault("types.problem_has_been_acknowledged", "string")
	viper.SetDefault("types.acknowledgement_type", "string")
	viper.SetDefault("types.current_notification_number", "string")
	viper.SetDefault("types.accept_passive_checks", "string")
	viper.SetDefault("types.event_handler_enabled", "string")
	viper.SetDefault("types.flap_detection_enabled", "string")
	viper.SetDefault("types.is_flapping", "string")
	viper.SetDefault("types.percent_state_change", "string")
	viper.SetDefault("types.latency", "string")
	viper.SetDefault("types.execution_time", "string")
	viper.SetDefault("types.scheduled_downtime_depth", "string")
	viper.SetDefault("types.process_performance_data", "string")
	viper.SetDefault("types.obsess", "string")

	if viper.InConfig("autoreload") && viper.GetBool("autoreload") {
		viper.WatchConfig()
		viper.OnConfigChange(func(in fsnotify.Event) {
			err := viper.ReadInConfig()
			if err != nil {
				log.Println("ERROR - Configuration changed, but it contains an error:\n", err.Error())
			} else {
				reloadTableHeader()
				log.Println("INFO - Configuration reloaded", in.String())
			}
		})
	}
	//viper.WriteConfig() // writes current config to predefined path set by 'viper.AddConfigPath()' and 'viper.SetConfigName'

	reloadTableHeader()
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

/*
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
}*/

func reloadTableHeader() {
	tableHeader = viper.GetStringSlice("tableHeaders")
	generateTableHeadersString()
}

func generateTableHeadersString() {
	columns := ""
	for _, h := range tableHeader {
		columns += `{"text":"` + h + `","type":"` + viper.GetString("types."+h) + `"},`
	}

	if len(columns) > 2 {
		columns = columns[:len(columns)-1]
	}
	tableHeaderString = columns

}
