package genchart

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
	"os"
	"time"
	
	"github.com/influxdata/influxdb/influxql"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"strings"
)

var colors1 = []string{`89da59`,`90afc5`,`375e97`,`ffbb00`,`5bc8ac`,`4cb5f5`,`6ab187`,`ec96a4`,`f0810f`,`f9a603`,`a1be95`,`e2dfa2`,`ebdf00`,`5b7065`,`eb8a3e`,`217ca3`}
var colors2 = []string{`4D4D4D`,`5DA5DA`,`FAA43A`,`60BD68`,`F17CB0`,`B2912F`,`B276B2`,`DECF3F`,`F15854`,`004185`,`099482`,`5df058`,`20b8ce`,`b9c1c9`,`e8d174`,`e39e54`,`d64d4d`,`4d7358`,`9ed670`}

const dbaddress = "http://localhost:8087"
const djson = `gd.json`

var timefilter string //= `time > now() - 2590h`

var c client.Client

func Genchart(tf string) {
    
    timefilter = tf
    
	title, chartnames, chartqueries, yaxisname := loaddashboards()
	c = newdbclient()

	var leftaxis, rightaxis string
	
	for i := 0; i < len(chartnames.Array()); i++ {

		if yaxisname.Array()[i].Array()[1].String() != "" {
			
			leftaxis = yaxisname.Array()[i].Array()[0].String()
			rightaxis = yaxisname.Array()[i].Array()[1].String()
			
			drawChart2(title.String(), leftaxis, rightaxis, chartnames.Array()[i].String(), chartqueries.Array()[i])
			//fmt.Println(title.String(), leftaxis, rightaxis, chartnames.Array()[i].String(), chartqueries.Array()[i])
			
		} else {
			
			leftaxis = yaxisname.Array()[i].Array()[0].String()
			
			drawChart1(title.String(), leftaxis, chartnames.Array()[i].String(), chartqueries.Array()[i])
			//fmt.Println(title.String(), leftaxis, chartnames.Array()[i].String(), chartqueries.Array()[i])
		}

	}

	c.Close()

}
func drawChart1(title, leftaxis, chartname string, chartqueries gjson.Result) {

	vts := make([]time.Time, 0)

	var cn1 []string

	qs := chartqueries.Array()[0]

	q1 := qs.Array()[0].String()

	cn, src, sql := parsesql(q1)

	cn1 = append(cn1, cn...)

	vts, yv1 := query(sql, cn, src)

	createChart1(title, leftaxis, chartname, cn1, vts, yv1)
}

func createChart1(title, leftaxis, chartname string, cn1 []string, vts []time.Time, yv1 [][]float64) {

	var min, max float64
	for ch := 0; ch < len(yv1); ch++ {

		if ch == 0 {
			min = yv1[ch][0]
			max = yv1[ch][0]
		}

		for _, v := range yv1[ch] {
			if v > max {
				max = v
			}
			if v < min {
				min = v
			}
		}
	}

	charts := make([]chart.Series, 0)

	for ch := 0; ch < len(yv1); ch++ {
		charts = append(charts, genchart1(`(R) `+cn1[ch+1], vts, yv1[ch], ch))
		fmt.Println("chart: ", ch)
	}

	graph := chart.Chart{
		Width:  1480,
		Height: 720,
		//Width:  1920, //1280,
		//Height: 1080, //720,
		DPI: 110,

		Background: chart.Style{
			Padding: chart.Box{
				Top:  20,
				Left: 270,
			},
		},

		XAxis: chart.XAxis{
			Name:      "time",
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),

			GridMajorStyle: chart.Style{
				Show:        true,
				StrokeColor: chart.ColorAlternateGray,
				StrokeWidth: 1.0,
			},

			ValueFormatter: chart.TimeMinuteValueFormatter, //TimeHourValueFormatter,
		},

		YAxis: chart.YAxis{
			Name:      leftaxis,
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),

			Range: &chart.ContinuousRange{
				Min: min - 0.05, //0.0,
				Max: max + 0.05, //10.0,
			},

		},

		Series: charts,
	}

	graph.Elements = []chart.Renderable{
		chart.LegendLeft(&graph),
	}

	buffer := bytes.NewBuffer([]byte{})

	err := graph.Render(chart.PNG, buffer)
	if err != nil {
		println("error on render chart ", err)
	}

	fo, err := os.Create(`images/` + chartname + ".png")
	if err != nil {
		panic(err)
	}

	if _, err := fo.Write(buffer.Bytes()); err != nil {
		panic(err)
	}

}

func drawChart2(title, leftaxis, rightaxis, chartname string, chartqueries gjson.Result) {

	vts := make([]time.Time, 0)

	var cn1, cn2 []string

	qs := chartqueries.Array()[0]

	q1 := qs.Array()[0].String()

	cn, src, sql := parsesql(q1)

	cn1 = append(cn1, cn...)

	vts, yv1 := query(sql, cn, src)

	q2 := qs.Array()[1].String()

	cn, src, sql = parsesql(q2)

	cn2 = append(cn2, cn...)

	vts, yv2 := query(sql, cn, src)

	createChart2(title, leftaxis, rightaxis, chartname, cn1, cn2, vts, yv1, yv2)
}

func createChart2(title, leftaxis, rightaxis, chartname string, cn1, cn2 []string, vts []time.Time, yv1, yv2 [][]float64) {

	var min1, max1 float64
	for ch := 0; ch < len(yv1); ch++ {

		if ch == 0 {
			min1 = yv1[ch][0]
			max1 = yv1[ch][0]
		}

		for _, v := range yv1[ch] {
			if v > max1 {
				max1 = v
			}
			if v < min1 {
				min1 = v
			}
		}
	}

	var min2, max2 float64
	for ch := 0; ch < len(yv2); ch++ {

		if ch == 0 {
			min2 = yv2[ch][0]
			max2 = yv2[ch][0]
		}

		for _, v := range yv2[ch] {
			if v > max2 {
				max2 = v
			}
			if v < min2 {
				min2 = v
			}
		}
	}

	charts := make([]chart.Series, 0)

	for ch := 0; ch < len(yv1); ch++ {
		charts = append(charts, genchart1(`(R) `+cn1[ch+1], vts, yv1[ch], ch))
		fmt.Println("chart y1: ", ch)
	}

	for ch := 0; ch < len(yv2); ch++ {
		charts = append(charts, genchart2(cn2[ch+1], vts, yv2[ch], ch))
		fmt.Println("chart y2: ", ch)
	}

	graph := chart.Chart{
		Width:  1480,
		Height: 720,
		DPI:    110,

		Background: chart.Style{
			Padding: chart.Box{
				Top:  20,
				Left: 270,
			},
		},

		XAxis: chart.XAxis{
			Name:      "time",
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),

			GridMajorStyle: chart.Style{
				Show:        true,
				StrokeColor: chart.ColorAlternateGray,
				StrokeWidth: 1.0,
			},

			ValueFormatter: chart.TimeMinuteValueFormatter,
		},

		YAxis: chart.YAxis{
			Name:      rightaxis,
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),

			Range: &chart.ContinuousRange{
				Min: min1 + 0.05,
				Max: max1 + 0.05,
			},
		},

		YAxisSecondary: chart.YAxis{
			Name:      leftaxis,
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),

			Range: &chart.ContinuousRange{
				Min: min2 + 0.05,
				Max: max2 + 0.05,
			},
		},

		Series: charts,
	}

	graph.Elements = []chart.Renderable{
		chart.LegendLeft(&graph),
	}

	buffer := bytes.NewBuffer([]byte{})

	err := graph.Render(chart.PNG, buffer)
	if err != nil {
		fmt.Println("error on render chart ", err)
	}

	fo, err := os.Create(`images/` + chartname + ".png")
	if err != nil {
		panic(err)
	}

	if _, err := fo.Write(buffer.Bytes()); err != nil {
		panic(err)
	}

}

func genchart1(column string, ts []time.Time, yv []float64, ch int) chart.TimeSeries {

	gc := chart.TimeSeries{
		Name: column,
		Style: chart.Style{
			Show:        true,
			StrokeWidth: 3,
			DotWidth:    2,
			DotColor:    drawing.ColorFromHex(colors1[ch]),
			StrokeColor: drawing.ColorFromHex(colors1[ch]),
		},
		
		XValues: ts,
		YValues: yv,
	}

	return gc

}

func genchart2(column string, ts []time.Time, yv []float64, ch int) chart.TimeSeries {

	gc := chart.TimeSeries{
		Name: column,
		Style: chart.Style{
			Show:        true,
			StrokeWidth: 3,
			DotWidth:    2,
			DotColor:    drawing.ColorFromHex(colors2[ch]),
			StrokeColor: drawing.ColorFromHex(colors2[ch]),
		},

		YAxis:   chart.YAxisSecondary,
		XValues: ts,
		YValues: yv,
	}

	return gc

}

func parsesql(sql string) ([]string, []string, string) {

	sql = strings.Replace(sql, "$timeFilter", timefilter, -1)
	stmt, err := influxql.NewParser(strings.NewReader(sql)).ParseStatement()

	if err != nil {
		println("error: ", err)
	}

	cn := stmt.(*influxql.SelectStatement).ColumnNames()

	src := stmt.(*influxql.SelectStatement).Sources.Names()

	return cn, src, sql

}

func query(sql string, cn, src []string) ([]time.Time, [][]float64) {
	
	var table string
	
	for _, s := range src {
		table = s
		println("sources: ", s)
	}

	for _, c := range cn {
		println("colnames: ", c)
	}

	q := client.NewQuery(sql, "telem_" + table, "s")

	vts := make([]time.Time, 0)

	vvs := make([][]float64, 0)

	if response, err := c.Query(q); err == nil && response.Error() == nil {

		for r := range response.Results {

			for s := range response.Results[r].Series {

				//fmt.Println("series name", r, ": ", response.Results[r].Series[s].Name)

				vals := response.Results[r].Series[s].Values

				for v := 0; v < len(vals); v++ { 				// time
					
					t, _ := vals[v][0].(json.Number).Int64()
					vts = append(vts, time.Unix(t, 0))					
					
					}

				vs := make([]float64, 0)
				
				for vv := 1; vv < len(vals[0]); vv++ { 		// number of columns returned minus time column
					
					for v := 0; v < len(vals); v++ {
						
						switch value := vals[v][vv].(type) {
							case string:

							case json.Number:
								
								f, err := value.Float64()
								if err != nil {
									println(err)
								}
								
								vs = append(vs, f)
							case bool:

							case nil:
								println("sql", sql)
								println("nil? ", "------------------NO DATA--------------")
								
								return []time.Time{}, [][]float64{}

							default:
								println("default?")

							}
					
				}
					
				vvs = append(vvs, vs)
				vs = make([]float64, 0)
				//fmt.Println("vs:", len(vals[0]), vv, vs)
				
				}
				
			}

		}

		//fmt.Println(vts)
		//fmt.Println(vvs)

	}
	
	return vts, vvs
}
	
func loaddashboards() (gjson.Result, gjson.Result, gjson.Result, gjson.Result) {
	content, err := ioutil.ReadFile(djson)
	
    if err!=nil{
        fmt.Print("Error:",err)
    }

	results := gjson.GetManyBytes(content, "title", "rows.#.panels.0.title", "rows.#.panels.#.targets.#.query", "rows.#.panels.0.yaxes.#.label")

    return results[0], results[1], results[2], results[3]

	}	

func newdbclient() client.Client {
	
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: dbaddress,
		})
		
	if err != nil {
		fmt.Println("Error creating InfluxDB Client: ", err.Error())
	}
	
	return c
	
	
	}

