package exception

func LogIfError(err error) {
	if err != nil {
		panic(err)
	}
}
