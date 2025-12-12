SHELL := /bin/bash
build:
	mkdir -p execs
	for exec in $$(go tool dist list \
		| grep -E 'windows|darwin|bsd|linux' \
		| grep -Ev '386|mips|ppc|loong64|s390x'); do \
		os=$${exec%%/*}; \
		arch=$${exec#*/}; \
		suffix=""; \
		if [[ $$os == "windows" ]]; then \
			suffix=".exe"; \
		fi; \
		echo "Building for $$os with $$arch"; \
		GOOS=$$os GOARCH=$$arch go build -ldflags="-s -w" \
			-o="execs/$$os-$$arch-3xp1$$suffix"; \
	done
clean:
	mkdir -p execs
	rm execs/*
