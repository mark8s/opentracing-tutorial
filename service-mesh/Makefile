.PHONY: build
build: summer-camp reading details

.PHONY: summer-camp
summer-camp:
	@docker build -t $(REGISTRY)/summer-camp-service:$(TAG) -f summer-camp/Dockerfile .

.PHONY: reading
reading:
	@docker build -t $(REGISTRY)/reading:$(TAG) -f reading/Dockerfile .

.PHONY: details
details:
	@docker build -t $(REGISTRY)/details:$(TAG) -f details/Dockerfile .

.PHONY: push
push: push-summer-camp push-reading push-details

.PHONY: push-summer-camp
push-summer-camp:
	@docker push $(REGISTRY)/summer-camp-service:$(TAG)

.PHONY: push-reading
push-reading:
	@docker push $(REGISTRY)/reading:$(TAG)

.PHONY: push-details
push-details:
	@docker push $(REGISTRY)/details:$(TAG)
