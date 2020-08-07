all:
	cd src && go build -o ../bin/monmop
install:
	cd src && go install
clean:
	@rm monmop
