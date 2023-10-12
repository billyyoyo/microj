package db

import (
	"fmt"
	"github.com/billyyoyo/microj/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
	"time"
)

type BaseModel struct {
	ID        int64          `gorm:"primarykey;column:id;type:bigint;"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

type User struct {
	BaseModel
	LoginName string `gorm:"column:login_name;type:varchar(64);uniqueIndex;not null;comment:''"`
	UserName  string `gorm:"column:user_name;type:varchar(64);not null"`
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
	if err := db2.Ping(); err == nil {
		fmt.Println("Connect success")
	} else {
		fmt.Println(err.Error())
	}
}
