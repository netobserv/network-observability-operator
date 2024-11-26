package utils

type tcpFlag struct {
	value uint16
	name  string
}

var tcpFlags = []tcpFlag{
	{value: 1, name: "FIN"},
	{value: 2, name: "SYN"},
	{value: 4, name: "RST"},
	{value: 8, name: "PSH"},
	{value: 16, name: "ACK"},
	{value: 32, name: "URG"},
	{value: 64, name: "ECE"},
	{value: 128, name: "CWR"},
	{value: 256, name: "SYN_ACK"},
	{value: 512, name: "FIN_ACK"},
	{value: 1024, name: "RST_ACK"},
}

func DecodeTCPFlags(bitfield uint16) []string {
	var values []string
	for _, flag := range tcpFlags {
		if bitfield&flag.value != 0 {
			values = append(values, flag.name)
		}
	}
	return values
}
