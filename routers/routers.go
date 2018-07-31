package routers

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	recaptcha "github.com/dpapathanasiou/go-recaptcha"
	humanize "github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
	"github.com/patrickmn/go-cache"
	"github.com/recoilme/tgram/models"
	"golang.org/x/text/language"
	"gopkg.in/russross/blackfriday.v2"
)

var (
	NBSecretPassword = "A String Very Very Very Niubilty!!@##$!@#4"
	ReCaptcha        = ""
	cc               *cache.Cache
)

const (
	CookieTime  = 2592000
	RateIP      = 10 * time.Minute
	RatePost    = 10 * time.Minute
	RateComment = 1 * time.Minute
)

func init() {
	cc = cache.New(24*time.Hour, 10*time.Minute)
}

// CheckAuth - general hook sets all param like lang, user
func CheckAuth() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Set("path", c.Request.URL.Path)
		var lang = "ru"
		var tokenStr, username, image string

		hosts := strings.Split(c.Request.Host, ".")
		var host = hosts[0]
		if host == "localhost:8081" {
			// dev
			c.Redirect(http.StatusFound, "http://sub."+host)
			return
		}
		if host == "tgr" {
			// tgr.am
			//fmt.Println("host:tgr")
			t, _, err := language.ParseAcceptLanguage(c.Request.Header.Get("Accept-Language"))
			if err == nil && len(t) > 0 {
				if len(t[0].String()) >= 2 {
					// some lang found
					if len(t[0].String()) == 2 || len(t[0].String()) == 3 {
						// some lang 3 char
						lang = t[0].String()
					} else {
						// remove country code en-US
						langs := strings.Split(t[0].String(), "-")
						if len(langs[0]) == 2 || len(langs[0]) == 3 {
							lang = langs[0]
						}
					}
				}
			}
			// redirect on subdomain
			c.Redirect(http.StatusFound, "http://"+lang+".tgr.am")
			return
		}
		if len(host) < 2 || len(host) > 3 {
			c.Redirect(http.StatusFound, "http://"+lang+".tgr.am")
			return
		}

		// store subdomain
		c.Set("lang", host)

		//fmt.Println("lang:", lang, "host:", host, "path", c.Request.URL.Path)

		// token from cookie
		if tokenС, err := c.Cookie("token"); err == nil && tokenС != "" {
			tokenStr = tokenС
		}

		if tokenStr == "" {
			// token from header
			authStr := c.Request.Header.Get("Authorization")
			// Strips 'TOKEN ' prefix from token string
			if len(authStr) > 5 && strings.ToUpper(authStr[0:6]) == "TOKEN " {
				tokenStr = authStr[6:]
			}
		}
		if tokenStr != "" {
			token, tokenErr := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				// Don't forget to validate the alg is what you expect:
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
				}

				// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
				return []byte(NBSecretPassword), nil
			})
			if tokenErr == nil && token != nil {
				//token found
				if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
					username = claims["username"].(string)
					image = claims["image"].(string)
				}
			}
		}
		c.Set("username", username)
		c.Set("image", image)
	}
}

// ToStr convert object to string
func ToStr(value interface{}) string {
	return fmt.Sprintf("%s", value)
}

// ToDate convert time 2 date
func ToDate(t time.Time) string {
	return fmt.Sprintf("%s", humanize.Time(t))
}

// GetLead return first paragraph
func GetLead(s string) string {
	var delim = strings.IndexRune(s, '\n')
	//log.Println(s, delim)
	if delim > 300 || delim < 0 {
		var len = len([]rune(s))
		if len > 300 {
			len = 300
		}
		delim = strings.LastIndexByte(s[:len], ' ')
		if delim < 0 {
			delim = len
		}
	}
	//log.Println(s, delim)
	return s[:delim] + ".."
}

// Home - main page
func Home(c *gin.Context) {

	username := c.GetString("username")
	var users []models.User
	if username != "" {
		users = models.IFollow(c.GetString("lang"), "fol", username)
	}
	c.Set("users", users)

	c.HTML(http.StatusOK, "home.html", c.Keys)
}

// All all page
func All(c *gin.Context) {
	articles, page, prev, next, last, err := models.AllArticles(c.GetString("lang"), c.Query("p"))
	if err != nil {
		renderErr(c, err)
		return
	}
	//log.Println(len(articles))
	c.Set("articles", articles)
	c.Set("page", page)
	c.Set("prev", prev)
	c.Set("next", next)
	c.Set("last", last)
	from_int, _ := strconv.Atoi(c.Query("p"))
	c.Set("p", from_int)
	c.HTML(http.StatusOK, "all.html", c.Keys)
}

func renderErr(c *gin.Context, err error) {
	switch c.Request.Header.Get("Accept") {
	case "application/json":
		// Respond with JSON
		c.JSON(http.StatusUnprocessableEntity, err)
	default:
		// Respond with HTML
		c.Set("err", err)
		c.HTML(http.StatusBadRequest, "err.html", c.Keys)
	}
}

func genToken(username, image string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"image":    image,
	})

	// Sign and get the complete encoded token as a string using the secret
	return token.SignedString([]byte(NBSecretPassword))
}

// Register page
func Register(c *gin.Context) {
	var err error
	switch c.Request.Method {
	case "GET":
		c.HTML(http.StatusOK, "register.html", c.Keys)
	case "POST":
		ip := c.ClientIP()
		//log.Println("ReCaptcha", ReCaptcha)
		ReCaptcha = "" // disable
		if ReCaptcha != "" {
			//validate if set
			c.Request.ParseForm()
			//_ = c.Request.PostFormValue("g-recaptcha-response")
			//log.Println("g-recaptcha-response", c.Request.PostFormValue("g-recaptcha-response"))
			recaptchaResponse, responseFound := c.Request.Form["g-recaptcha-response"]
			if responseFound {
				result, err := recaptcha.Confirm(ip, recaptchaResponse[0])
				//log.Println("recaptchaResponse", result, err)
				if err != nil {
					renderErr(c, err)
					return
				}
				if !result {
					renderErr(c, errors.New("Error validating captcha"))
					return
				}
			}
			wait := ratelimit(ip, RateIP)
			if wait > 0 {
				e := fmt.Sprintf("Rate limit on registraton from that ip, please wait: %d Seconds", wait)
				renderErr(c, errors.New(e))
				return
			}

		}
		var u models.User
		err = c.ShouldBind(&u)
		if err != nil {
			renderErr(c, err)
			return
		}

		// create user
		u.Username = strings.ToLower(u.Username)
		u.Lang = c.GetString("lang")
		u.IP = ip
		err = models.UserNew(&u)
		if err != nil {
			renderErr(c, err)
			return
		}

		tokenString, err := genToken(u.Username, "")
		if err != nil {
			renderErr(c, err)
			return
		}
		// add to cache on success
		cc.Set(ip, time.Now().Unix(), cache.DefaultExpiration)

		c.SetCookie("token", tokenString, CookieTime, "/", "", false, true)
		switch c.Request.Header.Get("Accept") {
		case "application/json":
			c.JSON(http.StatusCreated, u)
			return
		default:
			c.Redirect(http.StatusFound, "/")
		}

	default:
		c.HTML(http.StatusNotFound, "err.html", errors.New("not found"))
	}
}

// Settings page
func Settings(c *gin.Context) {

	switch c.Request.Method {
	case "GET":
		user, err := models.UserGet(c.GetString("lang"), c.GetString("username"))
		if err != nil {
			renderErr(c, err)
			return
		}
		c.Set("bio", user.Bio)
		c.Set("email", user.Email)
		c.Set("image", user.Image)
		c.HTML(http.StatusOK, "settings.html", c.Keys)
	case "POST":
		var u models.User
		var err error
		u.Username = c.GetString("username")
		err = c.ShouldBind(&u)
		if err != nil {
			renderErr(c, err)
			return
		}
		u.Lang = c.GetString("lang")
		user, err := models.UserCheckGet(u.Lang, u.Username, u.Password)
		if err != nil {
			renderErr(c, err)
			return
		}
		//fmt.Printf("user:%+v\n", u)
		u.Password = ""
		u.PasswordHash = user.PasswordHash
		err = models.UserSave(&u)
		if err != nil {
			renderErr(c, err)
			return
		}
		if u.Image != user.Image {
			// upd token
			tokenString, err := genToken(u.Username, u.Image)
			if err != nil {
				renderErr(c, err)
				return
			}
			if c.Request.Header.Get("Accept") != "application/json" {
				c.SetCookie("token", tokenString, CookieTime, "/", "", false, true)
			}

		}
		switch c.Request.Header.Get("Accept") {
		case "application/json":
			c.JSON(http.StatusOK, u)
			return
		default:
			c.Redirect(http.StatusFound, "/")
		}
	}
}

// Logout remove cookie
func Logout(c *gin.Context) {
	switch c.Request.Method {
	case "GET":
		c.SetCookie("token", "", 0, "/", "", false, true)
		c.Redirect(http.StatusFound, "/")
	case "POST":
	}
}

// Login page
func Login(c *gin.Context) {
	var err error
	switch c.Request.Method {
	case "GET":
		c.HTML(http.StatusOK, "login.html", c.Keys)
	case "POST":
		var u models.User
		err = c.ShouldBind(&u)
		if err != nil {
			renderErr(c, err)
			return
		}

		user, err := models.UserCheckGet(c.GetString("lang"), u.Username, u.Password)
		if err != nil {
			renderErr(c, err)
			return
		}
		tokenString, err := genToken(user.Username, user.Image)
		if err != nil {
			renderErr(c, err)
			return
		}
		c.SetCookie("token", tokenString, CookieTime, "/", "", false, true)
		switch c.Request.Header.Get("Accept") {
		case "application/json":
			c.JSON(http.StatusOK, user)
		default:
			c.Redirect(http.StatusFound, "/")
		}

	}
}

func ratelimit(key string, dur time.Duration) (wait int) {
	if x, found := cc.Get(key); found {
		// if found
		t := time.Now()
		elapsed := t.Sub(time.Unix(x.(int64), 0))

		if elapsed < dur {
			wait = int((dur - elapsed).Seconds())
		}
	}
	return wait
}

// Editor page
func Editor(c *gin.Context) {
	aid, _ := strconv.Atoi(c.Param("aid"))
	c.Set("aid", aid)
	postRate := c.GetString("lang") + ":p:" + c.GetString("username")
	switch c.Request.Method {
	case "GET":

		if aid > 0 {
			// check username
			username := c.GetString("username")
			a, err := models.ArticleGet(c.GetString("lang"), username, uint32(aid))
			if err != nil {
				renderErr(c, err)
				return
			}
			str := strings.Replace(a.Body, "\n\n", "\r\n", -1)
			c.Set("body", str)
			c.Set("title", a.Title)
			c.Set("ogimage", a.OgImage)
		} else {
			wait := ratelimit(postRate, RatePost)
			if wait > 0 {
				e := fmt.Sprintf("Rate limit for new users on new post, please wait: %d Seconds", wait)
				renderErr(c, errors.New(e))
				return
			}
		}
		c.HTML(http.StatusOK, "article_edit.html", c.Keys)
	case "POST":
		//log.Println("aid", aid)
		var err error
		var abind models.Article
		err = c.ShouldBind(&abind)
		if err != nil {
			renderErr(c, err)
			return
		}
		username := c.GetString("username")
		lang := c.GetString("lang")
		host := "http://" + c.Request.Host + "/"

		body, err := models.ImgProcess(strings.Replace(strings.TrimSpace(abind.Body), "\r\n", "\n\n", -1), lang, username, host)
		if err != nil {
			renderErr(c, err)
			return
		}

		unsafe := blackfriday.Run([]byte(body))
		html := template.HTML(bluemonday.UGCPolicy().SanitizeBytes(unsafe))

		title := abind.Title
		ogimage := abind.OgImage
		//log.Printf("ogimage:'%s'\n", ogimage)
		var a models.Article
		if aid > 0 {

			a, err := models.ArticleGet(lang, username, uint32(aid))
			if err != nil {
				renderErr(c, err)
				return
			}
			a.HTML = html
			a.Body = body
			a.Title = title
			a.OgImage = ogimage
			err = models.ArticleUpd(a)
			if err != nil {
				renderErr(c, err)
				return
			}
			//log.Println("aid2", a)
			//log.Println("Author", a.Author, "a.ID", a.ID, fmt.Sprintf("/@%s/%d", a.Author, a.ID))
			c.Redirect(http.StatusFound, fmt.Sprintf("/@%s/%d", a.Author, a.ID))
			return
		}
		wait := ratelimit(postRate, RatePost)
		if wait > 0 {
			e := fmt.Sprintf("Rate limit for new users on new post, please wait: %d Seconds", wait)
			renderErr(c, errors.New(e))
			return
		}
		if _, bannedAuthor := cc.Get("ban:uid:" + username); bannedAuthor {
			renderErr(c, errors.New("You are banned on 24 h for spam, advertising, illegal and / or copyrighted content. Sorry about that("))
			return
		}
		/*
			if _, bannedIP := cc.Get(c.ClientIP()); bannedIP {
				renderErr(c, errors.New("This ip was banned on 24 h for spam, advertising, illegal and / or copyrighted content. Sorry about that("))
				return
			}*/
		a.Lang = lang
		a.Author = username
		a.Image = c.GetString("image")
		a.OgImage = ogimage
		a.CreatedAt = time.Now()
		a.HTML = html
		a.Body = body
		a.Title = title
		newaid, err := models.ArticleNew(&a)
		if err != nil {
			renderErr(c, err)
			return
		}
		a.ID = newaid
		// add to cache on success
		cc.Set(postRate, time.Now().Unix(), cache.DefaultExpiration)

		//log.Println("Author", a.Author, "a.ID", a.ID, fmt.Sprintf("/@%s/%d", a.Author, a.ID))
		c.Redirect(http.StatusFound, fmt.Sprintf("/@%s/%d", a.Author, a.ID))
		return
	}
}

// Article page
func Article(c *gin.Context) {
	switch c.Request.Method {
	case "GET":
		lang := c.GetString("lang")
		aid, _ := strconv.Atoi(c.Param("aid"))
		aid32 := models.Uint32toBin(uint32(aid))
		username := c.Param("username")
		a, err := models.ArticleGet(lang, username, uint32(aid))
		if err != nil {
			renderErr(c, err)
			return
		}
		c.Set("link", "http://"+c.Request.Host+c.GetString("path"))
		c.Set("article", a)
		c.Set("title", a.Title)
		delim := strings.IndexByte(a.Body, '\n')
		if delim > 10 && delim < 300 {
			c.Set("description", a.Body[:delim])
		}
		//log.Printf("HTML:'%s'\n,", a.Body)
		c.Set("body", a.HTML)
		isFolow := models.IsFollowing(lang, "fol", username, c.GetString("username"))
		c.Set("isfollow", isFolow)
		followcnt := models.FollowCount(lang, "fol", username)
		c.Set("followcnt", followcnt)

		// fav
		isFav := models.IsFollowing(lang, "fav", string(aid32), c.GetString("username"))
		c.Set("isfav", isFav)
		favcnt := models.FollowCount(lang, "fav", string(aid32))
		c.Set("favcnt", favcnt)
		//log.Println("Art", a)

		// view counter
		unic := fmt.Sprintf("%s%d%s", lang, a.ID, c.ClientIP())
		unicCnt := fmt.Sprintf("%s%d", lang, a.ID)
		var view int
		if _, found := cc.Get(unic); !found {
			// new unique view for last 24 h - increment
			cc.Set(unic, 0, cache.DefaultExpiration) //store on 24 h
			v, notfounderr := cc.IncrementInt(unicCnt, 1)
			if notfounderr != nil {
				stored := models.ViewGet(lang, a.ID)
				cc.Add(unicCnt, stored, cache.NoExpiration)
				view = 1
			} else {
				view = v
				//if v%5 == 0 {
				models.ViewSet(lang, a.ID, v)
				//}
			}
		} else {
			if x, f := cc.Get(unicCnt); f {
				view = x.(int)
			}
		}
		c.Set("view", view)
		c.Set("ogimage", a.OgImage)
		c.HTML(http.StatusOK, "article.html", c.Keys)
		//c.JSON(http.StatusOK, a)
	}
}

// ArticleDelete delete page by id of current user
func ArticleDelete(c *gin.Context) {
	switch c.Request.Method {
	case "GET":
		aid, _ := strconv.Atoi(c.Param("aid"))
		username := c.GetString("username")
		err := models.ArticleDelete(c.GetString("lang"), username, uint32(aid))
		if err != nil {
			renderErr(c, err)
			return
		}
		// remove rate limit on delete
		cc.Delete(c.GetString("lang") + ":p:" + username)
		c.Redirect(http.StatusFound, "/")
	}
}

// Follow subscribe on user
func Follow(c *gin.Context) {
	switch c.Request.Method {
	case "GET":
		user := c.Param("user")
		username := c.GetString("username")
		action := c.Param("action")
		err := models.Following(c.GetString("lang"), "fol", user, username)
		if err != nil {
			renderErr(c, err)
			return
		}
		c.Redirect(http.StatusFound, action)
	}
}

// Unfollow unsubscribe
func Unfollow(c *gin.Context) {
	switch c.Request.Method {
	case "GET":
		user := c.Param("user")
		action := c.Param("action")
		err := models.Unfollowing(c.GetString("lang"), "fol", user, c.GetString("username"))
		if err != nil {
			renderErr(c, err)
			return
		}
		c.Redirect(http.StatusFound, action)
	}
}

// Fav add to favorites
func Fav(c *gin.Context) {
	switch c.Request.Method {
	case "GET":

		aid, _ := strconv.Atoi(c.Param("aid"))
		aid32 := models.Uint32toBin(uint32(aid))
		//fmt.Println(aid32, string(aid32), []byte(string(aid32)))
		action := c.Param("action")
		username := c.GetString("username")

		err := models.Following(c.GetString("lang"), "fav", string(aid32), username)
		if err != nil {
			renderErr(c, err)
			return
		}
		c.Redirect(http.StatusFound, action)
	}
}

// Unfav remove from favorites
func Unfav(c *gin.Context) {
	switch c.Request.Method {
	case "GET":
		aid, _ := strconv.Atoi(c.Param("aid"))
		aid32 := models.Uint32toBin(uint32(aid))
		action := c.Param("action")
		err := models.Unfollowing(c.GetString("lang"), "fav", string(aid32), c.GetString("username"))
		if err != nil {
			renderErr(c, err)
			return
		}
		c.Redirect(http.StatusFound, action)
	}
}

// GoToRegister redirect to registration
func GoToRegister() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString("username") == "" {
			c.Redirect(http.StatusFound, "/register")
			c.Abort()
		}
	}
}

// Author page
func Author(c *gin.Context) {
	authorStr := c.Param("username")
	lang := c.GetString("lang")
	author, err := models.UserGet(lang, authorStr)
	if err != nil {
		renderErr(c, err)
		return
	}

	articles, page, prev, next, last, err := models.ArticlesAuthor(c.GetString("lang"), c.GetString("username"), authorStr, c.Query("p"))
	if err != nil {
		renderErr(c, err)
		return
	}

	c.Set("articles", articles)
	c.Set("page", page)
	c.Set("prev", prev)
	c.Set("next", next)
	c.Set("last", last)
	from_int, _ := strconv.Atoi(c.Query("p"))
	c.Set("p", from_int)

	c.Set("author", author)
	isFolow := models.IsFollowing(lang, "fol", authorStr, c.GetString("username"))
	c.Set("isfollow", isFolow)
	followcnt := models.FollowCount(lang, "fol", authorStr)
	c.Set("followcnt", followcnt)

	// fav
	aid, _ := strconv.Atoi(c.Param("aid"))
	aid32 := models.Uint32toBin(uint32(aid))
	isFav := models.IsFollowing(lang, "fav", string(aid32), c.GetString("username"))
	c.Set("isfav", isFav)
	favcnt := models.FollowCount(lang, "fav", string(aid32))
	c.Set("favcnt", favcnt)
	c.HTML(http.StatusOK, "author.html", c.Keys)
}

// CommentNew create comment
func CommentNew(c *gin.Context) {
	switch c.Request.Method {
	case "POST":
		lang := c.GetString("lang")
		aid, _ := strconv.Atoi(c.Param("aid"))
		username := c.Param("username")

		var err error
		var a models.Article
		err = c.ShouldBind(&a)
		if err != nil {
			renderErr(c, err)
			return
		}
		rateComKey := lang + ":c:" + c.GetString("username")
		wait := ratelimit(rateComKey, RateComment)
		if wait > 0 {
			e := fmt.Sprintf("Rate limit for new users on new comment, please wait: %d Seconds", wait)
			renderErr(c, errors.New(e))
			return
		}
		//Replace http://sub.localhost:8081/@recoilme with @recoilme
		/*host := "http://" + c.Request.Host + "/@"
		a.Body = strings.Replace(a.Body, host, "@", -1)
		hosts := "https://" + c.Request.Host + "/@"
		a.Body = strings.Replace(a.Body, hosts, "@", -1)*/

		parsed, parsedUsers := models.Mention(strings.Replace(a.Body, "\r\n", "\n\n", -1), lang)
		_ = parsedUsers
		a.Body = parsed
		//log.Println("bod", a.Body)
		unsafe := blackfriday.Run([]byte(a.Body))
		html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
		a.HTML = template.HTML(html)

		a.Lang = lang
		a.Author = c.GetString("username")
		a.Image = c.GetString("image")
		a.CreatedAt = time.Now()

		//a.Body = string(body)
		cid, err := models.CommentNew(&a, username, uint32(aid))
		if err != nil {
			renderErr(c, err)
			return
		}

		// add to cache on success
		cc.Set(rateComKey, time.Now().Unix(), cache.DefaultExpiration)
		c.Redirect(http.StatusFound, fmt.Sprintf("/@%s/%d#comment%d", username, aid, cid))
		//c.JSON(http.StatusCreated, a) //gin.H{"article": serializer.Response()})
		//c.Redirect(http.Sta
	}
}

// Favorites return last 100 fav
func Favorites(c *gin.Context) {
	switch c.Request.Method {
	case "GET":
		lang := c.GetString("lang")
		user := c.Param("username")
		articles := models.Favorites(lang, user)
		c.Set("articles", articles)
		var prev, next, last uint32
		page := ""
		c.Set("page", page)
		c.Set("prev", prev)
		c.Set("next", next)
		c.Set("last", last)
		//from_int, _ := strconv.Atoi(c.Query("p"))
		from_int := 0
		c.Set("p", from_int)
		c.HTML(http.StatusOK, "all.html", c.Keys)
	}
}

// ArticleBad - delete article and ban author
func ArticleBad(c *gin.Context) {
	switch c.Request.Method {
	case "GET":
		aid, _ := strconv.Atoi(c.Param("aid"))
		author := c.Param("author")
		username := c.GetString("username")

		//check for me
		if username != "recoilme" {
			renderErr(c, errors.New("You are not recoilme"))
			return
		}
		// check for not me
		if author == "recoilme" {
			renderErr(c, errors.New("You are recoilme!"))
			return
		}
		err := models.ArticleDelete(c.GetString("lang"), author, uint32(aid))
		if err != nil {
			renderErr(c, err)
			return
		}
		// remove rate limit on delete
		cc.Delete(c.GetString("lang") + ":p:" + username)
		// add author ip to ban
		//_, errauthor := models.UserGet(c.GetString("lang"), author)
		//if errauthor == nil {
		/*if u.IP != "" {
			cc.Set(u.IP, time.Now().Unix(), cache.DefaultExpiration)
		}*/
		cc.Set("ban:uid:"+author, time.Now().Unix(), cache.DefaultExpiration)
		//}
		c.Redirect(http.StatusFound, "/@"+author)
	}
}

// Upload image
func Upload(c *gin.Context) {
	switch c.Request.Method {
	case "GET":
		c.HTML(http.StatusOK, "upload.html", c.Keys)
	case "POST":
		var fileHeader *multipart.FileHeader
		var err error
		var src multipart.File
		var newElement = ""
		minSize := 102400
		if fileHeader, err = c.FormFile("file"); err != nil {
			renderErr(c, err)
			return
		}
		//log.Println("Size:", fileHeader.Size)
		if fileHeader.Size > int64(minSize*100) {
			renderErr(c, errors.New("File too big"))
			return
		}

		if src, err = fileHeader.Open(); err == nil {
			defer src.Close()
			//log.Println("src:", src)
			b, err := ioutil.ReadAll(src)
			//log.Println("b:", len(b))
			if err != nil {
				renderErr(c, err)
				return
			}
			file, orig, origSize := models.Store("", c.GetString("lang"), c.GetString("username"), b)
			if file == "" || orig == "" {
				renderErr(c, errors.New("Some error"))
				return
			}
			//TODO https
			host := "http://" + c.Request.Host + "/"

			if origSize > minSize {
				newElement = "[![](" + host + file + " )](" + host + orig + ")"
			} else {
				newElement = "![](" + host + orig + ")"
			}

			c.Set("newelement", newElement)
			c.HTML(http.StatusOK, "upload.html", c.Keys)

		} else {
			//log.Println("open error", err)
			renderErr(c, err)
			return
		}

	}
}
