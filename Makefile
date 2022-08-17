build:
	docker build -t mermaid .
run:
	docker run --rm -it --name mermaid -p 8080:8080 mermaid