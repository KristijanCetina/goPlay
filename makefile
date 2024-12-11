.PHONY: all build clean run
# .DEFAULT_GOAL := build
# MAKEFLAGS += --silent

OUTPUT = main.out
# CXXFLAGS = -O3 -std=c++17 -Wall -Wextra -o $(OUTPUT)
FILES =  main.go
HEADERS = 

$(OUTPUT): $(FILES) $(HEADERS)
	go build -o $(OUTPUT) $(FILES)

build: $(FILES)
	go build -o $(OUTPUT) $(FILES) 

run: $(OUTPUT)
	./$(OUTPUT)

clean:
	@$(RM) *.out
	@$(RM) *.o
	@$(RM) __debug_bin*

%: %.go
	go build -o $@.out $^
	