package model

import "gorm.io/gorm"

// DB 全局数据库实例，由 bootstrap 初始化后赋值
var DB *gorm.DB
