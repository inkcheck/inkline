# inkline. Run `make` or `make help` for the list of targets.

.DEFAULT_GOAL := help
.PHONY: help build test diagrams site serve notices notary-setup release-check release-snapshot release clean

help: ## Show this help
	@awk 'BEGIN { FS = ":.*## " } \
		/^[a-zA-Z0-9_-]+:.*## / { printf "  \033[36m%-17s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

build: ## Build bin/inkline
	go build -o bin/inkline ./cmd/inkline

test: ## Run the tests
	go test ./...

diagrams: ## Render every example, light and dark, into the docs site
	./scripts/render-site-diagrams.sh

site: diagrams ## Build the docs site into site/public
	cd site && hugo --gc --minify

serve: diagrams ## Serve the docs site on http://localhost:1414/inkline/
	cd site && hugo server --port 1414


notices: ## Regenerate THIRD_PARTY_NOTICES.md after dependency or asset changes
	./scripts/notices.sh

notary-setup: ## Store notarisation credentials (Apple ID, app-specific password) in the keychain
	./scripts/notary-setup.sh

release-check: ## Check everything a release needs, without releasing
	./scripts/release.sh check

# GoReleaser refuses to run with more than one forge token set.
release-snapshot: ## Build release archives locally in dist/; no signing, nothing published
	env -u GITLAB_TOKEN -u GITEA_TOKEN GITHUB_TOKEN="$${GITHUB_TOKEN:-$$(gh auth token)}" \
		goreleaser release --snapshot --clean

release: ## Sign, notarise and publish the tagged HEAD to GitHub and the Homebrew tap
	./scripts/release.sh publish

clean: ## Remove build output
	rm -rf bin dist site/public site/resources site/static/diagrams
