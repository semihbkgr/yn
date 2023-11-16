build:
	go build .

run: build
	./yn < ./yaml/testdata/data.yaml
