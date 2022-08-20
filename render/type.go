/**
 * @Author: zzy
 * @Email: zhangzhongyuan@didiglobal.com
 * @Description:
 * @File: type.go
 * @Package: render
 * @Version: 1.0.0
 * @Date: 2022/8/20 16:29
 */

package render

// LinkTarget 链接打开目标类型
type LinkTarget string

// 链接打开目标
const (
	LinkTargetSelf   LinkTarget = "_self"
	LinkTargetBlank  LinkTarget = "_blank"
	LinkTargetParent LinkTarget = "_parent"
	LinkTargetTop    LinkTarget = "_top"
)
