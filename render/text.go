/**
 * @Author: zzy
 * @Email: zhangzhongyuan@didiglobal.com
 * @Description:
 * @File: text.go
 * @Package: render
 * @Version: 1.0.0
 * @Date: 2022/8/20 16:05
 */

package render

// TextSnippetType 文本片段类型
type TextSnippetType string

// TextSnippetStyle 文本片段样式
type TextSnippetStyle string

// 文本类型
const (
	TextSnippetTypeText TextSnippetType = "text" // 文本
	TextSnippetTypeIcon TextSnippetType = "icon" // 图标
)

// 文本样式
const (
	TextSnippetStyleSecondary TextSnippetStyle = "secondary"
	TextSnippetStyleSuccess   TextSnippetStyle = "success"
	TextSnippetStyleWarning   TextSnippetStyle = "warning"
	TextSnippetStyleDanger    TextSnippetStyle = "danger"
)

// TextSnippet 文本片段
type TextSnippet struct {
	Type   TextSnippetType  `json:"type"`             // 类型（默认text）
	Text   string           `json:"text,omitempty"`   // 文本
	Icon   string           `json:"icon,omitempty"`   // 图标
	Link   string           `json:"link,omitempty"`   // 链接
	Style  TextSnippetStyle `json:"style,omitempty"`  // 样式
	Target LinkTarget       `json:"target,omitempty"` // 目标
}

// TextLine 文本行
type TextLine []TextSnippet
