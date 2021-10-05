package main

import (
	"github.com/epowsal/orderfile"
)

// 一个名加数据键值数数据库：name_attributes
// id找name键值数据库：id_name
// 一个文章标题索引数据库：title-key_ids
// 一个文章内容索引数据库:content-key_ids
// 一个属性索引数据库:attrisent_ids

var name_attributes *orderfile.OrderFile
var id_name *orderfile.OrderFile
var attrisentskey_ids *orderfile.OrderFile
var nameskey_ids *orderfile.OrderFile

var id_title *orderfile.OrderFile
var titleskey_ids *orderfile.OrderFile
var contentskey_ids *orderfile.OrderFile
