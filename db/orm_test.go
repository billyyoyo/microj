package db

import (
	"fmt"
	"github.com/billyyoyo/microj/config"
	"gorm.io/driver/mysql"
	"testing"
)

type User struct {
	BaseModel
	LoginName string `gorm:"column:login_name;type:varchar(64);uniqueIndex;not null;comment:''"`
	UserName  string `gorm:"column:user_name;type:varchar(64);not null"`
	Face      string `gorm:"column:face;type:varchar(128);"`
	DeptId    int64  `gorm:"column:dept_id;type:bigint;"`
	TenantId  int64  `gorm:"column:tenant_id;type:bigint;"`
}

type Department struct {
	BaseModel
	Name     string `gorm:"column:name;type:varchar(64);"`
	ParentId int64  `gorm:"column:parent_id;type:bigint;"`
	Code     string `gorm:"column:code;type:varchar(128);"`
	Level    int8   `gorm:"column:level;type:tinyint"`
	Position int8   `gorm:"column:position;type:tinyint;"`
	TenantId string `gorm:"column:tenant_id;type:bigint;"`
}

type Tenant struct {
	BaseModel
	Name string `gorm:"column:name;unique"`
}

func TestConnect(t *testing.T) {
	config.Init()
	models := []interface{}{&User{}} //, &Department{}, &Tenant{}}
	InitDataSource(mysql.Open, models)
	fmt.Println(NextID())
	fmt.Println(NextID())
	fmt.Println(NextID())
	db2 := Orm()
	sdb, err := db2.DB()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if err = sdb.Ping(); err == nil {
		fmt.Println("Connect success")
	} else {
		fmt.Println(err.Error())
	}
}
