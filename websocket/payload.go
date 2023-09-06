package websocket

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/29 18:23
  @describe :
*/
// Payload 用于websocket的推送数据结构
type Payload struct {
	// Type 类型
	Type string `json:"type"`

	// Data 具体数据
	Data any `json:"data"`
}
