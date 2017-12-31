
# Edit SPECNAME for the name your spec
SPECNAME := disco
# Ensure SPECTOOLS points at your spectools
PANDOC := $(SPECTOOLS)/pandoc
CITEPROC := $(SPECTOOLS)/pandoc-citeproc

# Use "make", "make html", "make pdf", or "make clean"
all: html pdf

html: output/$(SPECNAME).html

pdf: output/$(SPECNAME).pdf

output/$(SPECNAME).html: $(SPECNAME).md $(PANDOC)/template_pandoc.html $(PANDOC)/spec_markdown.css $(CITEPROC)/ieee-with-url.csl $(CITEPROC)/general.bib my.bib
	pandoc $(SPECNAME).md --standalone --toc \
	        --from markdown\
		--template $(PANDOC)/template_pandoc.html \
		--metadata=pdfn:$(SPECNAME).pdf \
		--css=spec_markdown.css \
		--filter pandoc-citeproc \
		--bibliography=$(CITEPROC)/general.bib \
		--bibliography=my.bib \
		--csl=$(CITEPROC)/ieee-with-url.csl \
		-o output/$(SPECNAME).html
	cp $(PANDOC)/spec_markdown.css output

output/$(SPECNAME).pdf: $(SPECNAME).md $(PANDOC)/template_pandoc.latex $(CITEPROC)/ieee-with-url.csl $(CITEPROC)/general.bib my.bib
	pandoc $(SPECNAME).md --standalone --toc \
	        --from markdown\
		--template $(PANDOC)/template_pandoc.latex \
		--filter pandoc-citeproc \
		--bibliography=$(CITEPROC)/general.bib \
		--bibliography=my.bib \
		--csl=$(CITEPROC)/ieee-with-url.csl \
		-o output/$(SPECNAME).pdf

clean:
	rm -f output/$(SPECNAME).html output/spec_markdown.css output/$(SPECNAME).pdf
