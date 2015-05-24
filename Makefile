BIN_NAME=muhkeat

build: *.go alastalon_salissa.txt
	go build -o ${BIN_NAME}

alastalon_salissa.txt:
	wget http://wunderdog.fi/static/alastalon_salissa.zip
	unzip alastalon_salissa.zip
	rm alastalon_salissa.zip
	touch $@

clean:
	rm -f ${BIN_NAME}

.PHONY: clean
