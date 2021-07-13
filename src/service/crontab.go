package service

import (
	"Yearning-go/src/lib"
	"Yearning-go/src/model"
	"Yearning-go/src/parser"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/robfig/cron"
)

func StartCrontab() {
	CreateCrontab()
	//createTask()

}

func CreateCrontab() {
	cron2 := cron.New() //创建一个cron实例

	//执行定时任务（每5秒执行一次）
	err := cron2.AddFunc("*/5 * * * * *", createTask)
	if err != nil {
		fmt.Println(err)
	}

	//启动/关闭
	cron2.Start()
	defer cron2.Stop()
	select {
	//查询语句，保持程序运行，在这里等同于for{}
	}
}

func createTask() {
	model.DB().First(&model.GloPer)
	_ = json.Unmarshal(model.GloPer.Message, &model.GloMessage)
	_ = json.Unmarshal(model.GloPer.Ldap, &model.GloLdap)
	_ = json.Unmarshal(model.GloPer.Other, &model.GloOther)
	_ = json.Unmarshal(model.GloPer.AuditRole, &parser.FetchAuditRole)

	MsgLists := make(map[string][]map[string]string)

	err := GetQueryOrderLists(&MsgLists)
	if err != nil {
		return
	}

	err = GetSqlOrderLists(&MsgLists)
	if err != nil {
		return
	}

	for k, v := range MsgLists {
		if length := len(v); length != 0 {
			for i := 0; i < len(v); i++ {
				SendFeishu(k, v[i]["content"])
			}
		}
	}

}

// 获取查询列表
func GetQueryOrderLists(d *map[string][]map[string]string) (err error) {
	var lists []model.CoreQueryOrder
	msgLists := *d
	res := model.DB().Where("query_per=?", 2).Find(&lists)
	if res.Error != nil {
		return errors.New("aasdf")
	}

	if len(lists) != 0 {

		for _, info := range lists {
			var key string
			var content string
			if lib.TimeDifference(info.ExDate) {
				model.DB().Model(&model.CoreQueryOrder{}).Where("id=?", info.ID).Update(&model.CoreQueryOrder{QueryPer: 3})
				key = info.Username
				content = "你的查询请求已过期，请重新申请！"
			} else {
				key = info.Assigned
				content = info.Realname + "(" + info.Username + ")" + "正在发送查询请求\n申请说明：\n" + info.Text + "\n请及时审核处理！"
			}

			msg := make(map[string]string)
			msg["type"] = "1"
			msg["content"] = content

			slice2 := make([]map[string]string, 1)
			result, ok := msgLists[key]
			if !ok {
				slice2[0] = msg
			} else {
				slice2 = append(result, msg)
			}

			fmt.Println(slice2)
			msgLists[key] = slice2
		}
	}

	return
}

// 获取工单列表
func GetSqlOrderLists(d *map[string][]map[string]string) (err error) {
	var lists []model.CoreSqlOrder
	msgLists := *d
	res := model.DB().Where("status=?", "2").Find(&lists)

	if res.Error != nil {
		return errors.New(res.Error.Error())
	}

	if len(lists) != 0 {
		for _, info := range lists {
			var typeText string
			typeText = "DML"
			if info.Type == 2 {
				typeText = "DDL"
			}
			content := info.RealName + "(" + info.Username + ")" + "正在申请" + typeText + "操作\n申请说明：\n" + info.Text + "\n" + "SQL明细：\n" + info.SQL + "\n请及时审核处理！"
			msg := make(map[string]string, 1)
			msg["type"] = "2"
			msg["content"] = content

			slice3 := make([]map[string]string, 1)
			ress, ok := msgLists[info.Assigned]
			if !ok {
				slice3[0] = msg
			} else {
				slice3 = append(ress, msg)
			}

			msgLists[info.Assigned] = slice3
		}
	}

	return
}

func SendFeishu(toUser string, msg string) {
	fmt.Println(msg)
}
