package aws

const instanceRunning = 16

func instanceState(vp *int32) byte {
	if vp == nil {
		return 0
	}
	return byte((*vp) & 0xff)
}

func drefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
