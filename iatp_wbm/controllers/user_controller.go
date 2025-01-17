package controllers

import (
	"fmt"
	"iatp/common/domain"
	ldap_tool "iatp/common/ldap"
	"iatp/iatp_wbm/services"
	"iatp/setting"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"github.com/kataras/iris/v12/sessions"
)

type UserController struct {
	Ctx     iris.Context
	Service services.UserService
	Session *sessions.Session
}

// Method: Post
// Resiurce: /user/list
func (c *UserController) PostList() mvc.Result {
	if c.Session.Get("authenticated") == nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"status": 500,
				"msg":    "认证失效,需重新认证",
				"data": map[string]interface{}{
					"items": []map[string]interface{}{},
				},
			},
		}
	}

	name := c.Ctx.PostValue("user_name")

	search_result := c.Service.SearchByName(name)
	if search_result == nil {
		return mvc.Response{
			Code: 200,
		}
	}

	return mvc.Response{
		ContentType: "application/json",
		Object:      search_result,
	}
}

// Method: Post
// Resiurce: /user/create
func (c *UserController) PostCreate() mvc.Result {
	if c.Session.Get("authenticated") == nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"status": 500,
				"msg":    "认证失效,需重新认证",
				"data": map[string]interface{}{
					"items": []map[string]interface{}{},
				},
			},
		}
	}

	name := c.Ctx.PostValue("user_name")

	status := c.Service.InsertUser(name)

	if status {
		return mvc.Response{
			Code: 200,
		}
	}

	return mvc.Response{
		Code: 500,
	}

}

// Method: Post
// Resiurce: /user/delete
func (c *UserController) PostDelete() mvc.Result {
	if c.Session.Get("authenticated") == nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"status": 500,
				"msg":    "认证失效,需重新认证",
				"data": map[string]interface{}{
					"items": []map[string]interface{}{},
				},
			},
		}
	}

	name := c.Ctx.PostValue("user_name")

	status := c.Service.DeleteUser(name)
	if status {
		return mvc.Response{
			Code: 200,
		}
	}

	return mvc.Response{
		Code: 500,
	}

}

func (c *UserController) PostLogin() mvc.Result {
	user_name := c.Ctx.PostValue("user_name")
	password := c.Ctx.PostValue("password")

	// 判断用户账户是否注册
	result := c.Service.SearchByName(user_name)
	if result == nil {
		return mvc.Response{
			Code: 500,
			Object: map[string]interface{}{
				"status": 500,
				"msg":    "不存在该用户",
			},
		}
	}

	setting.GetAllSettings()
	d, _ := domain.NewDomain(setting.IatpSetting.ReadSet("auth_domain").(string))
	// TODO 新增
	// auth_client := ldap_tool.NewLdap(d.DomainServer, d.GetUserDN(user_name), password, d.GetDomainScope(), d.SSL)
	auth_client := ldap_tool.NewLdap(d.DomainServer, user_name, password, d.GetDomainScope(), d.SSL)

	login, _ := auth_client.CheckConn()

	if login {
		c.Session.Set("authenticated", true)
		entrys := auth_client.SearchEntryByCN(user_name, []string{"displayName"}, nil)
		c.Session.Set("user_name", entrys[0].GetAttributeValue("displayName"))
	} else {
		return mvc.Response{
			Code: 500,
			Object: map[string]interface{}{
				"status": 500,
				"msg":    "验证失败",
			},
		}
	}

	return mvc.Response{
		Code: 200,
		Object: map[string]interface{}{
			"status": 0,
			"msg":    "登录成功",
		},
	}
}

func (c *UserController) PostLogout() mvc.Result {
	c.Session.Set("authenticated", nil)
	return mvc.Response{
		Code: 0,
		Object: map[string]interface{}{
			"status": 0,
			"msg":    "",
			"data":   map[string]interface{}{},
		},
	}
}

func (c *UserController) GetCurrent() mvc.Response {
	if c.Session.Get("authenticated") == nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code":    0,
				"message": "请求成功",
				"data": map[string]interface{}{
					"user_name": "未登录",
				},
			},
		}
	}

	return mvc.Response{
		Object: map[string]interface{}{
			"code":    0,
			"message": "请求成功",
			"data": map[string]interface{}{
				"user_name": c.Session.GetString("user_name"),
			},
		},
	}
}

// Method: Post
// Resiurce: /user/activity
func (c *UserController) PostActivity() mvc.Response {
	if c.Session.Get("authenticated") == nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"status": 500,
				"msg":    "认证失效,需重新认证",
				"data": map[string]interface{}{
					"items": []map[string]interface{}{},
				},
			},
		}
	}

	input_datetime_range := c.Ctx.PostValueDefault("input-datetime-range", "")
	if input_datetime_range == "" {
		input_datetime_range = fmt.Sprintf("%v,%v", time.Now().Add(-7*24*time.Hour).Unix(), time.Now().Unix())
	}

	user_name := c.Ctx.PostValueDefault("activity_user_name", "")
	logon_source := c.Ctx.PostValueDefault("logon_source", "")
	if len(user_name) == 0 {
		user_name = "administrator"
	}

	activity := c.Service.SearchActivity(user_name, logon_source, input_datetime_range)

	var status int
	var msg string

	if len(activity) > 0 {
		status = 0
		msg = "查询成功"
	} else {
		status = 1
		msg = "暂无数据"
	}

	return mvc.Response{
		Code: 200,
		Object: map[string]interface{}{
			"status": status,
			"msg":    msg,
			"data": map[string]interface{}{
				"items": activity,
			},
		},
	}
}
