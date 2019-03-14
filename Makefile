deps:
	// not sure that we need to type it right here
	// alternativly user should type it in terminal before making deps cause of go modules
	export GO111MODULE=on
	go get -t -d ./...
