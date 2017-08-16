.PHONY: build
build:
	docker build -t stor.highloadcup.ru/travels/real_piranha .
	docker push stor.highloadcup.ru/travels/real_piranha
