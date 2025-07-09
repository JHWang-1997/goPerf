package model

// 获取校验器
func GetVerify(conf *PerfConf) Verify {
	if conf.VerifyByEvent != 0 {
		return func(report *Report) bool {
			return report.FollowTokenCount+1 == uint64(conf.VerifyByEvent)
		}
	}
	code := 200
	if conf.VerifyByCode != 0 {
		code = conf.VerifyByCode
	}
	return func(report *Report) bool {
		return code == report.StatCode
	}
}
