package main

import (
	"fmt"
	"github.com/djimenez/iconv-go"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	reg_url     = "http://bbs.zg163.net/bbs/register.php"
	vote_url    = "http://bbs.zg163.net/bbs/hack.php?"
	log_out_url = "http://bbs.zg163.net/bbs/login.php?action=quit&verify="
)

func init() {
	proxy := os.Getenv("sock_proxy")
	if proxy == "" {
		proxy = os.Getenv("SOCK_PROXY")
	}
	if len(proxy) < 5 {
		log.Println("未找到sock代理，将使用http代理(如设置)")
	} else {
		log.Println("使用sock代理", proxy)
		http.DefaultTransport = &http.Transport{Dial: func(tp, addr string) (c net.Conn, err error) {
			c, err = net.Dial(tp, proxy)
			if err != nil {
				log.Println("代理服务器连接失败，将直连")
				return net.Dial(tp, addr)
			} else {
				log.Println("使用代理服务器连接成功")
				return
			}
		}, Proxy: http.ProxyFromEnvironment}
	}
}

func main() {
	for {
		rsp := vote()
		if len(rsp) < 2000 {
			if strings.Index(rsp, "操作成功") > 0 {
				continue
			} else {
				break
			}
		} else {
			break
		}
	}
}

func vote() string {
	client := http.DefaultClient
	client.Jar = new(myJar)
	rsp, err := client.Get(reg_url)
	if err != nil {
		panic(err)
	}

	rspBytes, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		panic(err)
	}
	rsp.Body.Close()

	rspStr := string(rspBytes)
	verify := findValue(&rspStr, "var verifyhash = '")
	hexie := findValue(&rspStr, "document.register._hexie.value='")

	if len(verify) < 5 || len(hexie) < 5 {
		msg, _ := iconv.ConvertString("取回验证信息出错", "utf-8", "gb2312")
		panic(msg)
	} else {
		log.Println("取验证信息成功，进行注册")
	}

	client.Get(log_out_url + verify)

	rand.Seed(int64(time.Now().Nanosecond()))

	uname := fmt.Sprintf("%2da12", rand.Int())

	form := make(url.Values)
	form.Add("regname", uname)
	form.Add("forward", "")
	form.Add("step", "2")
	form.Add("_hexie", hexie)
	form.Add("regpwd", uname+"123a")
	form.Add("regpwdrepeat", uname+"123a")
	form.Add("regemail", uname+"123d@sina.com")
	form.Add("apartment", "120101")
	form.Add("question", "0")
	form.Add("customquest", "")
	form.Add("answer", "")
	form.Add("rgpermit", "1")
	now := time.Now()
	rsp, err = client.PostForm(fmt.Sprintf("%s?now=%d&verify=%s", reg_url, now.Unix(), verify), form)
	if err != nil {
		panic(err)
	}
	rspBytes, err = ioutil.ReadAll(rsp.Body)
	if err != nil {
		panic(err)
	}
	if len(rspBytes) > 1000 {
		log.Fatal("估计注册未成功,返回内容太长:" + string(rspBytes))
	} else {
		log.Println("成功注册帐号:" + uname + "，开始投票")
	}

	vForm := make(url.Values)
	vForm.Add("H_name", "ext_thread")
	vForm.Add("threadtype", "picvote")
	vForm.Add("ext_action", "vote")
	vForm.Add("tid", "2933886")
	vForm.Add("item_id", "507")
	vForm.Add("action", "ajax")
	vForm.Add("verify", verify)
	vForm.Add("noewtime", fmt.Sprint(now.Add(2*time.Minute).Unix()))
	//错误，您所在IP投票次数已满
	//登录
	rsp, err = client.Get(vote_url + vForm.Encode())
	rspBytes, _ = ioutil.ReadAll(rsp.Body)
	var gbkRsp string
	if len(rspBytes) < 1000 {
		gbkRsp, _ = iconv.ConvertString(string(rspBytes), "gbk", "utf-8")
	}else{
		gbkRsp = string(rspBytes)
	}

	log.Println("投票完成，返回内容为:", gbkRsp)
	return gbkRsp
}

func findValue(str *string, head string) string {
	start := strings.Index(*str, head)
	if start < 0 {
		log.Println("could find start for head:", head)
	}
	start = start + len(head)
	end := strings.Index((*str)[start+1:], "'")
	if end < 0 {
		log.Println("could find end for head:", head)
	}
	return (*str)[start : start+end+1]
}

type myJar struct {
	store map[string][]*http.Cookie
}

func (self *myJar) Cookies(u *url.URL) (c []*http.Cookie) {
	if self.store == nil {
		self.store = make(map[string][]*http.Cookie)
	}
	return self.store[u.Host]
}

func (self *myJar) SetCookies(u *url.URL, c []*http.Cookie) {
	if self.store == nil {
		self.store = make(map[string][]*http.Cookie)
	}
	self.store[u.Host] = c
}

var _ http.CookieJar = new(myJar)
