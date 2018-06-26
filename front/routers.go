package front

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/recoilme/tgram/articles"
	"github.com/recoilme/tgram/common"
	"github.com/recoilme/tgram/users"
)

// Index show main page
func Index(c *gin.Context) {
	var user users.UserModel
	iuser, uexists := c.Get("my_user_model")
	if uexists {
		user = iuser.(users.UserModel)
	}

	//c.gin.H["my_user_model"] = loggedInInterface.(userModelValidator.UserModel)
	renderTemplate(c, "index", gin.H{
		"my_user_model": user,
	})
}

// Register new user
func Register(c *gin.Context) {
	if c.Request.Method == "GET" {
		renderTemplate(c, "register", gin.H{})
	}
	if c.Request.Method == "POST" {
		userModelValidator := users.NewUserModelValidator()

		if err := userModelValidator.Bind(c); err != nil {
			c.HTML(http.StatusBadRequest, "register.html", gin.H{
				"ErrorTitle":   "Registration Failed",
				"ErrorMessage": err.Error(),
				"Username":     userModelValidator.User.Username,
				"Email":        userModelValidator.User.Email,
				"Password":     userModelValidator.User.Password})
			return
		}
		if err := users.SaveOne(&userModelValidator.UserModel); err != nil {
			c.HTML(http.StatusBadRequest, "register.html", gin.H{
				"ErrorTitle":   "database",
				"ErrorMessage": err.Error(),
				"Username":     userModelValidator.User.Username,
				"Email":        userModelValidator.User.Email,
				"Password":     userModelValidator.User.Password})
			return
		}
		c.Set("my_user_model", userModelValidator.UserModel)
		c.SetCookie("token", common.GenToken(userModelValidator.UserModel.ID), 3600, "/", "", false, true)
		c.Redirect(http.StatusFound, "/")
	}
}

// Login user
func Login(c *gin.Context) {
	if c.Request.Method == "GET" {
		renderTemplate(c, "login", gin.H{})
	}
	if c.Request.Method == "POST" {
		loginValidator := users.NewLoginValidator()
		if err := loginValidator.Bind(c); err != nil {
			c.HTML(http.StatusBadRequest, "login.html", gin.H{
				"ErrorTitle":   "Login Failed",
				"ErrorMessage": err.Error(),
				"Email":        loginValidator.User.Email,
				"Password":     loginValidator.User.Password})
			return
		}
		userModel, err := users.FindOneUser(&users.UserModel{Email: loginValidator.UserModel.Email})

		if err != nil {
			//c.JSON(http.StatusForbidden, common.NewError("login", errors.New("Not Registered email or invalid password")))
			c.HTML(http.StatusBadRequest, "login.html", gin.H{
				"ErrorTitle":   "Login Failed",
				"ErrorMessage": "Not Registered email or invalid password"})
			return
		}

		if userModel.CheckPassword(loginValidator.User.Password) != nil {
			c.HTML(http.StatusBadRequest, "login.html", gin.H{
				"ErrorTitle":   "Login Failed",
				"ErrorMessage": "Not Registered email or invalid password"})
			//c.JSON(http.StatusForbidden, common.NewError("login", errors.New("Not Registered email or invalid password")))
			return
		}
		users.UpdateContextUserModel(c, userModel.ID)
		c.SetCookie("token", common.GenToken(userModel.ID), 3600, "/", "", false, true)
		//serializer := UserSerializer{c}
		//c.JSON(http.StatusOK, gin.H{"user": serializer.Response()})

		//c.Set("my_user_model", userModelValidator.UserModel)
		c.Redirect(http.StatusFound, "/")
	}

}

// Logout clear cookie
func Logout(c *gin.Context) {
	c.SetCookie("token", "", 0, "/", "", false, true)
	c.Redirect(http.StatusFound, "/")
}

// Editor create new article
func Editor(c *gin.Context) {
	log.Println("Editor", c.Request.Method)
	if c.Request.Method == "GET" {
		renderTemplate(c, "article_edit", gin.H{
			"my_user_model": c.MustGet("my_user_model").(users.UserModel)})

	}
	if c.Request.Method == "POST" {
		articleModelValidator := articles.NewArticleModelValidator()
		if err := articleModelValidator.Bind(c); err != nil {
			//c.JSON(http.StatusUnprocessableEntity, common.NewValidatorError(err))

			c.HTML(http.StatusBadRequest, "article_edit.html", gin.H{
				"ErrorTitle":   "Post Failed",
				"ErrorMessage": err.Error(),
				"Title":        articleModelValidator.Article.Title,
				"Description":  articleModelValidator.Article.Description,
				"Body":         articleModelValidator.Article.Body,
				"Tags":         articleModelValidator.Article.Tags})
			return
		}

		if err := articles.SaveOne(&articleModelValidator.ArticleModel); err != nil {
			//c.JSON(http.StatusUnprocessableEntity, common.NewError("database", err))
			c.HTML(http.StatusBadRequest, "article_edit.html", gin.H{
				"ErrorTitle":   "Post Failed (database)",
				"ErrorMessage": err.Error()})
			return
		}
		//serializer := ArticleSerializer{c, articleModelValidator.articleModel}
		//c.JSON(http.StatusCreated, gin.H{"article": serializer.Response()})
		c.Redirect(http.StatusFound, "/article/"+articleModelValidator.ArticleModel.Slug)
	}
}

// Settings show settings
func Settings(c *gin.Context) {
	var user users.UserModel
	iuser, uexists := c.Get("my_user_model")
	if uexists {
		user = iuser.(users.UserModel)
		log.Println("UserModel:", user)
	}
	if c.Request.Method == "GET" {
		renderTemplate(c, "settings", gin.H{"my_user_model": user})
	}
	if c.Request.Method == "POST" {
		userModelValidator := users.NewUserModelValidatorFillWith(user)
		if err := userModelValidator.Bind(c); err != nil {
			c.HTML(http.StatusBadRequest, "settings.html", gin.H{
				"ErrorTitle":    "Update Failed",
				"ErrorMessage":  err.Error(),
				"my_user_model": user})
			return
		}

		userModelValidator.UserModel.ID = user.ID
		if err := user.Update(userModelValidator.UserModel); err != nil {
			c.HTML(http.StatusBadRequest, "settings.html", gin.H{
				"ErrorTitle":    "Update Failed Database",
				"ErrorMessage":  err.Error(),
				"my_user_model": user})
			return
		}
		users.UpdateContextUserModel(c, user.ID)
		c.Redirect(http.StatusFound, "/")
	}
}

func renderTemplate(c *gin.Context, tmpl string, p interface{}) {
	c.HTML(http.StatusOK, tmpl+".html", p)
}
