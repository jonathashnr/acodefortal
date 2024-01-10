build:
	./tailwindcss -i ./templates/input.css -o ./static/output.css --minify
	go build -o bin/ajudafortaleza

run: build
	./bin/ajudafortaleza

tailwindw:
	./tailwindcss -i ./templates/input.css -o ./static/output.css --watch

getTailwind:
	curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64
	chmod +x tailwindcss-linux-x64
	mv tailwindcss-linux-x64 tailwindcss