// 复现用临时文件残留（repro artifact），不参与构建逻辑，安全忽略。
// 环境禁止删除文件，故保留为占位；原始复现已验证 panic: index out of range [-1]。
package main
