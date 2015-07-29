# Builds static assets
# Depends on:
# - scss
# - coffeescript
# Run `make` to compile static assets and produce an executable binary which will start patchy
# Run `make run` to compile static assets and run tenshi

.PHONY: all static clean run

STYLES:=$(patsubst styles/%.scss,static/css/%.css,$(wildcard styles/*.scss))
STYLES+=$(patsubst styles/%.css,static/css/%.css,$(wildcard styles/*.css))
SCRIPTS:=$(patsubst scripts/%.coffee,static/js/%.js,$(wildcard scripts/*.coffee))
SCRIPTS+=$(patsubst scripts/%.js,static/js/%.js,$(wildcard scripts/*.js))
IMAGES:=$(patsubst image/%,static/image/%,$(wildcard image/*))
PAGES:=$(patsubst pages/%,static/%,$(wildcard pages/*))

static/image/%: image/%
	@mkdir -p static/image/
	cp $< $@

static/%: pages/%
	@mkdir -p static/
	cp $< $@

static/css/%.css: styles/%.css
	@mkdir -p static/css
	cp $< $@

static/css/%.css: styles/%.scss
	@mkdir -p static/css
	scss -I styles/ $< $@

static/js/%.js: scripts/%.js
	@mkdir -p static/js
	cp $< $@

static/js/%.js: scripts/%.coffee
	@mkdir -p static/js
	coffee -m -o static/ -c $<

static: $(STYLES) $(SCRIPTS) $(IMAGES) ${PAGES}
	@mkdir -p static/queue/
	@mkdir -p temp/
	go build -o patchy src/main.go src/hub.go src/conn.go src/transcoder.go src/songHandler.go src/util.go src/getHandling.go src/queue.go src/timer.go src/library.go

run: $(STYLES) $(SCRIPTS) $(IMAGES) ${PAGES}
	go run src/main.go src/hub.go src/conn.go src/transcoder.go src/mpd.go src/util.go src/getHandling.go src/queue.go src/timer.go src/library.go

all: static
	echo $(STYLES)
	echo $(SCRIPTS)
	echo ${IMAGES}
	echo ${PAGES}

clean:
	rm -rf static
