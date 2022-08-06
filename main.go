package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	http "github.com/useflyent/fhttp"
)

type Captcha struct {
	CaptchaSitekey string `json:"captcha_sitekey"`
	RqData         string `json:"captcha_rqdata"`
	RqToken        string `json:"captcha_rqtoken"`
}

type TaskPayload struct {
	ClientKey   string `json:"clientKey"`
	IsInvisible bool   `json:"isInvisible"`
	Data        string `json:"data"`
	UserAgent   string `json:"userAgent"`
	Task        struct {
		Type       string `json:"type"`
		WebsiteURL string `json:"websiteURL"`
		WebsiteKey string `json:"websiteKey"`
	} `json:"task"`
}

type TaskIdStruct struct {
	TaskId int `json:"taskId"`
}

type CreateTaskStruct struct {
	ClientKey string `json:"clientKey"`
	TaskId    int    `json:"taskId"`
}

type SolutionStruct struct {
	Status    string `json:"status"`
	ErrorCode string `json:"errorCode"`
	Solution  struct {
		CaptchaResponse string `json:"gRecaptchaResponse"`
	} `json:"solution"`
}

type VInvite struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Guild   struct {
		Name      string `json:"name"`
		VanityUrl string `json:"vanity_url_code"`
	} `json:"guild"`
}

var ValidatedInvs []string

func solveCaptcha(httpClient *http.Client, captcha Captcha, TaskId TaskIdStruct, Solution SolutionStruct) string {

	payload := TaskPayload{ClientKey: "57e79074192061f7e36a2eb5d6bed7a1", Data: captcha.RqData, IsInvisible: true, UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:102.0) Gecko/20100101 Firefox/102.0", Task: struct {
		Type       string "json:\"type\""
		WebsiteURL string "json:\"websiteURL\""
		WebsiteKey string "json:\"websiteKey\""
	}{Type: "HCaptchaTaskProxyless", WebsiteURL: "https://discord.com/", WebsiteKey: captcha.CaptchaSitekey}}
	payloadJSON, _ := json.Marshal(payload)

	taskId, _ := http.NewRequest("POST", "https://api.capmonster.cloud/createTask", bytes.NewBuffer(payloadJSON))
	taskId.Header.Set("content-type", "application/json")
	taskId.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:102.0) Gecko/20100101 Firefox/102.0")
	taskIdResp, _ := httpClient.Do(taskId)
	fuckyou, _ := ioutil.ReadAll(taskIdResp.Body)
	json.Unmarshal(fuckyou, &TaskId)

	println(TaskId.TaskId)

	createTask := CreateTaskStruct{ClientKey: "57e79074192061f7e36a2eb5d6bed7a1", TaskId: TaskId.TaskId}
	createTaskJson, _ := json.Marshal(createTask)

	for Solution.Solution.CaptchaResponse == "" && Solution.ErrorCode == "" {
		resp3, _ := http.Post("https://api.capmonster.cloud/getTaskResult", "application/json", bytes.NewBuffer(createTaskJson))
		yeah, _ := ioutil.ReadAll(resp3.Body)

		json.Unmarshal(yeah, &Solution)
		println(Solution.Status, Solution.ErrorCode)

		time.Sleep(3 * time.Second)
	}

	if Solution.ErrorCode != "" {
		switch Solution.ErrorCode {
		case "ERROR_ZERO_BALANCE":
			log.Fatal("CapMonster Balance Too Low")
			break

		case "ERROR_CAPTCHA_UNSOLVABLE":
			log.Fatal("Captcha Can Not Be Solved")
			break

		case "ERROR_IP_BANNED":
			log.Fatal("Your Ass Got Ip Banned From Capmonster, Bozo")
			break

		case "ERROR_DOMAIN_NOT_ALLOWED":
			log.Fatal("Can't Solve Captchas From This Domain")
			break

		case "ERROR_TOKEN_EXPIRED":
			log.Fatal("Your Capmonster Token Has Expired Loser")
			break
		}
	}

	return Solution.Solution.CaptchaResponse
}

func validateInvites(invs []string, httpClient *http.Client) []string {
	var validatedInvs []string

	for index := range invs {
		var Invite VInvite
		req, _ := http.NewRequest("GET", "https://discord.com/api/v9/invites/"+invs[index], nil)
		req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36")
		req.Header.Set("authorization", "OTk1Mjc5MTcwNDQxNTg0Njcw.GwMqvp.6eJD3oUmnXvErjEcuSEJ_eXk-NZ6HExSYYzA10")

		resp, _ := httpClient.Do(req)

		gamer, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(gamer, &Invite)

		//println(Invite.Code, Invite.Message, string(gamer))

		if Invite.Code != 10006 {
			validatedInvs = append(validatedInvs, Invite.Guild.VanityUrl)
		}
		time.Sleep(1 * time.Second)
	}

	return validatedInvs
}

func gatherInvs(httpClient *http.Client) {
	i := 1

	for len(ValidatedInvs) <= 1 {

		req, _ := http.NewRequest("GET", "https://discord.st/?page="+strconv.Itoa(i), nil)
		req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36")

		resp, err := httpClient.Do(req)

		if err != nil {
			panic(err)
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)

		if err != nil {
			log.Fatal(err)
		}

		var invs []string

		f := func(i int, s *goquery.Selection) bool {

			link, _ := s.Attr("href")
			return strings.HasPrefix(link, "/server")
		}

		doc.Find("body a").FilterFunction(f).Each(func(_ int, tag *goquery.Selection) {

			link, _ := tag.Attr("href")
			for i := 0; i < len(strings.Split(link, "/"))%3; i++ {
				test := strings.ReplaceAll(strings.Split(link, "/")[2], "-", "")
				invs = append(invs, test)
			}

		})

		validatedInvsTemp := validateInvites(invs, httpClient)

		ValidatedInvs = append(ValidatedInvs, validatedInvsTemp...)

		i++
	}
}

func main() {
	var captcha Captcha
	var TaskId TaskIdStruct
	var Solution SolutionStruct

	i := 0

	httpClient := &http.Client{}

	gatherInvs(httpClient)

	f, _ := os.Open("tokens.txt")
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		token := scanner.Text()
		what := strings.Split(token, ":")

		println(ValidatedInvs[i])

		req2, _ := http.NewRequest("POST", "https://discord.com/api/v9/invites/"+ValidatedInvs[i], bytes.NewReader([]byte(`{}`))) //+ValidatedInvs[i], nil)
		req2.Header.Set("Authorization", what[2])
		//req2.Header.Set("Host", "discord.com")
		//req2.Header.Set("Accept", "*/*")
		//req2.Header.Set("Accept-Language", "en-US,en;q=0.5")
		//req2.Header.Set("Accept-Encoding", "gzip, deflate, br")
		//req2.Header.Set("Content-Type", "application/json")
		//req2.Header.Set("X-Context-Properties", "eyJsb2NhdGlvbiI6Ikludml0ZSBCdXR0b24gRW1iZWQiLCJsb2NhdGlvbl9ndWlsZF9pZCI6bnVsbCwibG9jYXRpb25fY2hhbm5lbF9pZCI6IjEwMDUyMDAwNzM0NTU1NzkyNTciLCJsb2NhdGlvbl9jaGFubmVsX3R5cGUiOjEsImxvY2F0aW9uX21lc3NhZ2VfaWQiOiIxMDA1MjAwMTAxMzI2NzM3NDIyIn0=")
		//req2.Header.Set("Authorization", "OTk1NzA5MTAwNDg4OTIxMDk4.GLLQgl.Z8Qz7vhUjRhBemP3uDq8JEjgalh29PjaZG6gF0")
		req2.Header.Set("X-Super-Properties", "eyJvcyI6IldpbmRvd3MiLCJicm93c2VyIjoiRmlyZWZveCIsImRldmljZSI6IiIsInN5c3RlbV9sb2NhbGUiOiJlbi1VUyIsImJyb3dzZXJfdXNlcl9hZ2VudCI6Ik1vemlsbGEvNS4wIChXaW5kb3dzIE5UIDEwLjA7IFdpbjY0OyB4NjQ7IHJ2OjEwMi4wKSBHZWNrby8yMDEwMDEwMSBGaXJlZm94LzEwMi4wIiwiYnJvd3Nlcl92ZXJzaW9uIjoiMTAyLjAiLCJvc192ZXJzaW9uIjoiMTAiLCJyZWZlcnJlciI6IiIsInJlZmVycmluZ19kb21haW4iOiIiLCJyZWZlcnJlcl9jdXJyZW50IjoiIiwicmVmZXJyaW5nX2RvbWFpbl9jdXJyZW50IjoiIiwicmVsZWFzZV9jaGFubmVsIjoic3RhYmxlIiwiY2xpZW50X2J1aWxkX251bWJlciI6MTQwMzU1LCJjbGllbnRfZXZlbnRfc291cmNlIjpudWxsfQ==")
		//req2.Header.Set("X-Discord-Locale", "en-US")
		//req2.Header.Set("X-Debug-Options", "bugReporterEnabled")
		//req2.Header.Set("Content-Length", "2")
		//req2.Header.Set("Origin", "https://discord.com")
		//req2.Header.Set("DNT", "1")
		//req2.Header.Set("Alt-Used", "discord.com")
		//req2.Header.Set("Connection", "keep-alive")
		//req2.Header.Set("Referer", "https://discord.com/channels/@me/1005200073455579257")
		//req2.Header.Set("Cookie", "__dcfduid=af825bb014dc11edbacb35c6559f2a54; __sdcfduid=af825bb114dc11edbacb35c6559f2a54def7315ef4df20fed144a929bb2a23c7a442c518345f133f89909e32d138c7bd;")
		//req2.Header.Set("Sec-Fetch-Dest", "empty")
		//req2.Header.Set("Sec-Fetch-Mode", "cors")
		//req2.Header.Set("Sec-Fetch-Site", "same-origin")
		req2.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:102.0) Gecko/20100101 Firefox/102.0")
		req2.Header.Set("Referer", "https://discord.com/channels/@me/")

		resp2, err := httpClient.Do(req2)
		yess, _ := ioutil.ReadAll(resp2.Body)

		if err != nil {
			log.Fatal("Fatal error: " + err.Error())
		}

		json.Unmarshal(yess, &captcha)

		println(string(yess), resp2.Status)

		if captcha.CaptchaSitekey != "" {
			captchaAnswer := solveCaptcha(httpClient, captcha, TaskId, Solution)
			req3, _ := http.NewRequest("POST", "https://discord.com/api/v9/invites/"+ValidatedInvs[i], bytes.NewReader([]byte(fmt.Sprintf(`{"captcha_key":"%s","captcha_rqtoken":"%s"}`, captchaAnswer, captcha.RqToken))))
			req3.Header.Set("Authorization", what[2])
			req3.Header.Set("Content-type", "application/json")
			req3.Header.Set("User-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:102.0) Gecko/20100101 Firefox/102.0")
			req3.Header.Set("Referer", "https://discord.com/channels/@me/")
			req3.Header.Set("X-Super-Properties", "eyJvcyI6IldpbmRvd3MiLCJicm93c2VyIjoiRmlyZWZveCIsImRldmljZSI6IiIsInN5c3RlbV9sb2NhbGUiOiJlbi1VUyIsImJyb3dzZXJfdXNlcl9hZ2VudCI6Ik1vemlsbGEvNS4wIChXaW5kb3dzIE5UIDEwLjA7IFdpbjY0OyB4NjQ7IHJ2OjEwMi4wKSBHZWNrby8yMDEwMDEwMSBGaXJlZm94LzEwMi4wIiwiYnJvd3Nlcl92ZXJzaW9uIjoiMTAyLjAiLCJvc192ZXJzaW9uIjoiMTAiLCJyZWZlcnJlciI6IiIsInJlZmVycmluZ19kb21haW4iOiIiLCJyZWZlcnJlcl9jdXJyZW50IjoiIiwicmVmZXJyaW5nX2RvbWFpbl9jdXJyZW50IjoiIiwicmVsZWFzZV9jaGFubmVsIjoic3RhYmxlIiwiY2xpZW50X2J1aWxkX251bWJlciI6MTQwMzU1LCJjbGllbnRfZXZlbnRfc291cmNlIjpudWxsfQ==")
			//req2.Header.Set("cookie", "__dcfduid=af825bb014dc11edbacb35c6559f2a54; __sdcfduid=af825bb114dc11edbacb35c6559f2a54def7315ef4df20fed144a929bb2a23c7a442c518345f133f89909e32d138c7bd;")

			resp3, err := httpClient.Do(req3)
			yesss, _ := ioutil.ReadAll(resp3.Body)

			if err != nil {
				log.Fatal("Fatal error: " + err.Error())
			}

			json.Unmarshal(yesss, &captcha)

			println(string(yesss), resp3.Status)
		}

		time.Sleep(3 * time.Second)
		i++
		break
	}

}
