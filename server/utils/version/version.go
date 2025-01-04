package version

func LT(version1, version2 string) bool {
	if version2 == "test" {
		return true
	}

	return version1 < version2
}

func GT(version1, version2 string) bool {
	if version2 == "test" {
		return false
	}
	return version1 > version2
}
