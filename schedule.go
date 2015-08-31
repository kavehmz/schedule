package main

/*
This is a CLI tool to query channels schedule.
It is the way I find easier to skim.
Format of call is: schedule.go [today|tomorrow]
I found the API to be open to use:
http://api-epg.astro.com.my/ElasticEPGAPI_deploy/api/json/metadata?op=GuideRequest

This is also my first go app I wrote to learn how go works.
I liked it. := has a nostalgic touch.
*/

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"time"
	"unicode/utf8"
)

type Schedule struct {
	Start  string
	End    string
	Guides []Guide
}

type Guide struct {
	Event_id                     uint64
	Parental_rating_id           uint64
	Genre_title                  string
	Subgenre_title               string
	Display_datetime             string
	Display_datetime_end         string
	Display_datetime_end_utc     string
	Activation_datetime          string
	Display_datetime_utc         string
	Name                         string
	Description                  string
	Extended_description         string
	Actors                       string
	Producers                    string
	Transmission_type_premier    bool
	Transmission_type_repeat     bool
	Transmission_type_exhibition bool
	Transmission_type_fixed_time bool
	Subtitled                    bool
	Content                      string
	Service_title                string
	Nowshowing                   bool
	ProgramTitle                 string
	Custom_day                   string
}

func favorite_channels() [11][2]string {
	return [11][2]string{
		{"175", "431"},
		{"73", "412"},
		{"187", "443"},
		{"272", "434"},
		{"168", "571"},
		{"217", "430"},
		{"172", "573"},
		{"176", "575"},
		{"41", "554"},
		{"44", "556"},
		{"290", "725"},
	}
}

func get_coming_channel_schedule(channel [2]string) []Guide {
	var guides []Guide

	if len(os.Args) == 1 || os.Args[1] == "today" {
		var guides_today []Guide = get_schedule_text_for(channel[0], time.Now())
		for _, each := range guides_today {
			t, _ := time.Parse("2006-01-02 15:04:05", each.Display_datetime_end_utc)
			if t.Unix() >= time.Now().Unix() {
				each.Custom_day = "Today"
				guides = append(guides, each)
			}
		}
	}

	if len(os.Args) == 1 || os.Args[1] == "tomorrow" {
		var guides_tomorrow []Guide = get_schedule_text_for(channel[0], time.Now().Add(time.Hour*24))
		for _, each := range guides_tomorrow {
			each.Custom_day = "Tomorrow"
			guides = append(guides, each)
		}
	}

	return guides
}

func get_schedule_text_for(channel string, schedule_date time.Time) []Guide {
	resp, err := http.Get("http://api-epg.astro.com.my/api/guide/start/" + schedule_date.Format("2006-01-02") + "/end/" + schedule_date.Format("2006-01-02") + "T23:59:59/channels/" + channel + "?format=jsonp")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return parse_channel_schedule_text(string(body))
}

func parse_channel_schedule_text(content string) []Guide {
	var data Schedule
	var err = json.Unmarshal([]byte(content), &data)
	if err != nil {
		fmt.Println(err)
	}

	re := regexp.MustCompile("(.+)T(.+)\\.0000000Z$")
	for index, _ := range data.Guides {
		data.Guides[index].Display_datetime = re.ReplaceAllString(data.Guides[index].Display_datetime, "$1 $2")
		data.Guides[index].Display_datetime_utc = re.ReplaceAllString(data.Guides[index].Display_datetime_utc, "$1 $2")
		data.Guides[index].Display_datetime_end = re.ReplaceAllString(data.Guides[index].Display_datetime_end, "$1 $2")
		data.Guides[index].Display_datetime_end_utc = re.ReplaceAllString(data.Guides[index].Display_datetime_end_utc, "$1 $2")

	}

	return data.Guides
}

func program_description(guide Guide) string {
	var s string

	re_date := regexp.MustCompile("(.+) (.+)")
	re_desc := regexp.MustCompile("(.{0,90})\\b.*")
	s += "\x1b[31;1m" + guide.Name + "\x1b[0m : "
	s += "\x1b[34;1m" + guide.Custom_day + " ( "
	s += re_date.ReplaceAllString(guide.Display_datetime, "$2") + ": "
	s += re_date.ReplaceAllString(guide.Display_datetime_end, "$2") + " )\x1b[0m : "
	s += "\x1b[32;1m" + guide.Subgenre_title + "\x1b[0m : "
	s += re_desc.ReplaceAllString(guide.Description, "$1")
	return s

}

func show_channel_schedule(guides []Guide) {
	for _, each := range guides {
		fmt.Printf("%s\n", program_description(each))
	}
}

func show_channel_title(title string) {
	fmt.Println("\x1b[32;1mChannel: " + title)
	for i := 1; i < utf8.RuneCountInString(title)+10; i++ {
		fmt.Print("=")
	}
	fmt.Println("\x1b[0m")
}

func main() {

	for _, each := range favorite_channels() {
		var guides []Guide = get_coming_channel_schedule(each)
		if len(guides) == 0 {
			continue
		}
		show_channel_title(guides[0].Service_title)
		show_channel_schedule(guides)

	}

}
