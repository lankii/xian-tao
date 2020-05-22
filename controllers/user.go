package controllers

import (
	"encoding/base64"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/utils"
	"regexp"
	"strconv"
	"time"
	"xian-tao/models"
)

type UserController struct {
	beego.Controller
}

// 获取用户session
func GetUser(this *beego.Controller) string {
	userName := this.GetSession("userName")
	if userName != nil{
		this.Data["userName"] = userName.(string)
		return userName.(string)
	} else {
		this.Data["userName"] = ""
		return ""
	}
}

// 展示注册页面
func (this *UserController) ShowReg() {
	this.TplName = "register.html"
}

// 处理注册数据
func (this *UserController) HandleRed() {
	// 获取数据
	userName := this.GetString("user_name")
	password := this.GetString("pwd")
	confirmPassword := this.GetString("cpwd")
	email := this.GetString("email")

	// 校验数据
	// 判断是否为空
	if userName == "" || password == "" || confirmPassword == "" || email == "" {
		this.Data["errMsg"] = "数据填写不完整，请重新输入~"
		this.TplName = "register.html"
		return
	}
	// 判断密码与确认密码是否一致
	if password != confirmPassword {
		this.Data["errMsg"] = "两次密码输入的不一致，请重新输入~"
		this.TplName = "register.html"
		return
	}
	// 使用正则判断邮箱格式
	regex, _ := regexp.Compile("\\w[-\\w.+]*@([A-Za-z0-9][-A-Za-z0-9]+\\.)+[A-Za-z]{2,14}")
	resEmail := regex.FindString(email)
	if resEmail == "" {
		this.Data["errMsg"] = "邮箱格式不正确，请重新输入~"
		this.TplName = "register.html"
		return
	}

	// 处理数据
	o := orm.NewOrm()
	user := models.User{}
	user.Name = userName
	user.Password = password
	user.Email = email
	_, err := o.Insert(&user)
	if err != nil {
		this.Data["errMsg"] = "用户名已经存在，请重新输入~"
		this.TplName = "register.html"
		return
	}

	// 发送激活邮件
	emailConfig := `{"username":"431592976@qq.com","password":"nabvchjshgkacbbb","host":"smtp.qq.com","port":587}`
	emailConn := utils.NewEMail(emailConfig)
	emailConn.From = "431592976@qq.com" // 指定发件人的邮箱地址
	emailConn.To = []string{user.Email} // 指定收件人邮箱地址
	emailConn.Subject = "鲜淘驿站用户激活"      // 指定邮件的标题
	// 发给用户的是激活请求地址
	emailConn.HTML = `尊敬的` + user.Name + `，您好<br><br>感谢您注册鲜淘驿站，为了避免您忘记账号或密码导致您的账户无法找回，请您验证Email地址。<br><br>请复制粘贴下面的链接至浏览器地址栏打开：<br><br>127.0.0.1:8080/active?id=` + strconv.Itoa(user.Id) + `<br>`
	err = emailConn.Send()
	if err != nil {
		this.Data["errMsg"] = "发送激活邮件失败，请重新发送~"
		this.TplName = "register.html"
		return
	}

	// 返回视图
	this.Ctx.WriteString("注册成功，请去邮箱激活用户。")
	// c.Redirect("/login", 302)
}

// 激活用户
func (this *UserController) ActiveUser() {
	// 获取数据
	id, err := this.GetInt("id")

	// 校验数据
	if err != nil {
		this.Data["errMsg"] = "要激活的用户不存在"
		this.TplName = "register.html"
		return
	}
	// 处理数据
	o := orm.NewOrm()
	var user models.User
	user.Id = id
	err = o.Read(&user)
	if err != nil {
		this.Data["errMsg"] = "要激活的用户不存在"
		this.TplName = "register.html"
		return
	}
	user.Active = true
	o.Update(&user)
	// 返回视图
	this.Redirect("/login", 302)
}

// 展示登录页面
func (this *UserController) ShowLogin() {
	userName := this.Ctx.GetCookie("userName")
	// base64解密
	tempUserName, _ := base64.StdEncoding.DecodeString(userName)
	if string(tempUserName) == "" {
		// 没有记住用户名
		this.Data["userName"] = ""
		this.Data["checked"] = ""
	} else {
		this.Data["userName"] = string(tempUserName)
		this.Data["checked"] = "checked"
	}
	this.TplName = "login.html"
}

// 处理登录数据
func (this *UserController) HandleLogin() {
	// 获取数据
	userName := this.GetString("username")
	password := this.GetString("pwd")
	// 校验数据
	if userName == "" || password == "" {
		this.Data["errMsg"] = "用户名或密码不能为空"
		this.TplName = "login.html"
		return
	}

	// 处理数据
	o := orm.NewOrm()
	var user models.User
	user.Name = userName
	err := o.Read(&user, "Name")
	if err != nil {
		this.Data["errMsg"] = "用户名或密码错误"
		this.Data["userName"] = userName
		this.TplName = "login.html"
		return
	}
	if user.Password != password {
		this.Data["errMsg"] = "用户名或密码错误"
		this.Data["userName"] = userName
		this.TplName = "login.html"
		return
	}
	if user.Active != true {
		this.Data["errMsg"] = "用户名未激活，请先前往邮箱激活"
		this.Data["userName"] = userName
		this.TplName = "login.html"
		return
	}
	// 记住用户名处理
	rememberMe := this.GetString("remember_me")
	if rememberMe == "on" {
		// base64加密
		tempUserName := base64.StdEncoding.EncodeToString([]byte(userName))
		this.Ctx.SetCookie("userName", tempUserName, time.Second*3600*24*3)
	} else {
		this.Ctx.SetCookie("userName", userName, -1)
	}
	// 登录成功，设置session
	this.SetSession("userName", userName)

	// 返回视图
	this.Redirect("/", 302)
}

// 退出登录
func (this *UserController) Logout() {
	this.DelSession("userName")
	// 返回视图
	this.Redirect("/", 302)
}

// 展示用户中心信息页面
func (this *UserController) ShowUserInfo() {
	userName := GetUser(&this.Controller)

	// 查询地址表的内容
	o := orm.NewOrm()
	// 高级查询  表关联
	var addr models.Address
	o.QueryTable("Address").RelatedSel("User").Filter("User__Name", userName).Filter("IsDefault", true).One(&addr)
	if addr.Id == 0 {
		this.Data["addr"] = ""
	} else {
		this.Data["addr"] = addr
	}


	//c.Layout = "user_layout.html"
	this.Data["userName"] = userName
	this.TplName = "user_center_info.html"
}

// 展示用户中心订单页面
func (this *UserController) ShowUserOrder(){
	GetUser(&this.Controller)
	this.TplName = "user_center_order.html"
}

// 展示用户中心收货地址页面
func (this *UserController) ShowUserAddress(){
	userName := GetUser(&this.Controller)
	o := orm.NewOrm()
	var addr models.Address
	o.QueryTable("Address").RelatedSel("User").Filter("User__Name", userName).Filter("IsDefault", true).One(&addr)

	this.Data["addr"] = addr
	this.TplName = "user_center_site.html"
}

// 处理用户中心收货地址数据
func (this *UserController) HandleUserAddress(){
	// 获取数据
	receiver := this.GetString("receiver")
	address := this.GetString("address")
	zipCode := this.GetString("zip_code")
	phone := this.GetString("phone")
	// 校验数据(推荐使用正则校验，这里先使用判空)
	if receiver == "" || address == "" || zipCode == "" || phone == "" {
		this.Data["errMsg"] = "添加数据不完整，请重新输入~"
		this.Data["receiver"] = receiver
		this.Data["address"] = address
		this.Data["zipCode"] = zipCode
		this.Data["phone"] = phone
		this.TplName = "user_center_site.html"
		return
	}
	// 处理数据(插入数据)
	o := orm.NewOrm()
	var userAddress models.Address
	userAddress.IsDefault = true
	err := o.Read(&userAddress, "IsDefault")
	// 添加默认地址之前需要把原来的默认地址更改成非默认地址
	if err == nil {
		userAddress.IsDefault = false
		o.Update(&userAddress)
	}
	/*	更新默认地址时，给原来的地址对象的ID赋值了
	这时用原来的地址对象插入意思是用原来的ID做插入操作，会报错。*/
	// 关联用户表
	userName := this.GetSession("userName")
	var user models.User
	user.Name = userName.(string)
	o.Read(&user, "Name")

	var newUserAddress models.Address
	newUserAddress.Receiver = receiver
	newUserAddress.Addr = address
	newUserAddress.ZipCode = zipCode
	newUserAddress.Phone = phone
	newUserAddress.IsDefault = true
	newUserAddress.User = &user
	o.Insert(&newUserAddress)

	// 返回视图
	this.Redirect("/u/address", 302)
}
