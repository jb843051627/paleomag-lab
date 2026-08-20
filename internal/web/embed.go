package web

import "embed"

// FS 包含实验工作台的静态资源。
//
//go:embed static/*
var FS embed.FS
