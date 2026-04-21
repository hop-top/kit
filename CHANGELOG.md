# Changelog

## [0.1.0-alpha.0](https://github.com/hop-top/kit/compare/kit/v0.0.1-alpha.0...kit/v0.1.0-alpha.0) (2026-04-21)


### ⚠ BREAKING CHANGES

* open sourcing

### Miscellaneous

* open sourcing ([21e1b76](https://github.com/hop-top/kit/commit/21e1b76da40af4a8ee55aeae1efe41403c800256))

## [0.3.3](https://github.com/hop-top/kit/compare/kit-v0.3.2...kit-v0.3.3) (2026-04-18)

### Features

* **api:** HTTP API framework with auto-CRUD ([#24](https://github.com/hop-top/kit/issues/24)) ([2a4a49d](https://github.com/hop-top/kit/commit/2a4a49dd7c157918156f174eb4026ffc293a25e1))
* **ash:** agnostic session handler — full stack ([#31](https://github.com/hop-top/kit/issues/31)) ([bd7b797](https://github.com/hop-top/kit/commit/bd7b797053a50bb88112cad2b959b1b8244fbf26))
* **bus:** in-process pub/sub for CLI hooking ([#19](https://github.com/hop-top/kit/issues/19)) ([bf1eda0](https://github.com/hop-top/kit/commit/bf1eda03a659f7f47411f37c03e3e60d36aba93c))
* **bus:** pluggable adapter interface + SQLite cross-process adapter ([851316e](https://github.com/hop-top/kit/commit/851316ef576c24515960439b753cbef93f03a4e9))
* charm.land v2 CLI foundation ([8e87ae8](https://github.com/hop-top/kit/commit/8e87ae8d42f273b4b48fc048fa31f0762f9fc2ba))
* **ci:** changelog rewrite for release-please PRs ([0a4ca05](https://github.com/hop-top/kit/commit/0a4ca056433bcb70a15b94066a19f8855f3abf17))
* **ci:** reusable cli-demo-media workflow ([144e49c](https://github.com/hop-top/kit/commit/144e49c0f390df10ea0001526a301175aea0c651))
* **ci:** unified release pipeline via release-please ([6c2a209](https://github.com/hop-top/kit/commit/6c2a2096a0fbb2baef566f700e21033602edecbd))
* **cli:** add opinionated cobra+viper CLI setup ([d159c55](https://github.com/hop-top/kit/commit/d159c55c45abdadfac3fdf809d8f1c989c4b2c18))
* **cli:** add ResetFlags to prevent flag state leakage ([#34](https://github.com/hop-top/kit/issues/34)) ([c1636df](https://github.com/hop-top/kit/commit/c1636df44328b996b4d728cc20fd900f9d34392b))
* **cli:** command aliasing ([#12](https://github.com/hop-top/kit/issues/12)) ([7a0522e](https://github.com/hop-top/kit/commit/7a0522ed233562906f66604a1937188b2d45d014))
* **cli:** cross-language CLI/TUI parity — Go, TypeScript, Python ([60001a4](https://github.com/hop-top/kit/commit/60001a44dfae2dbfcd149d714b35a8c1237ffe4b))
* **cli:** git-style external plugin dispatch ([#20](https://github.com/hop-top/kit/issues/20)) ([512277c](https://github.com/hop-top/kit/commit/512277c86476ba7c508f1eeec1aed9aeb9da10ea))
* **cli:** rewrite with fang + charmtone theme ([4db2955](https://github.com/hop-top/kit/commit/4db29555f36921a8ac45f0ba33ff9fe1d90198f8))
* **config:** add layered config loader ([23d09bf](https://github.com/hop-top/kit/commit/23d09bf439e5a1a25044825e08eee50206c656c9))
* **config:** git-style get/set + PKL schema + wizard pipeline ([#39](https://github.com/hop-top/kit/issues/39)) ([60e6b0f](https://github.com/hop-top/kit/commit/60e6b0ff0a3d96d93f28f7033b040fee54cc5803))
* **devcontainer:** nix-based dev environment with Docker and AI tools ([4cb8f59](https://github.com/hop-top/kit/commit/4cb8f59674ec4bee4187b045d88665669e3f1f13)), closes [#7](https://github.com/hop-top/kit/issues/7)
* **domain/sqlite:** support composite primary keys ([#25](https://github.com/hop-top/kit/issues/25)) ([dc2a4f8](https://github.com/hop-top/kit/commit/dc2a4f83950e1faba1fe43faa5f3e820985d6c8a))
* **domain:** enrich transition errors with allowed states ([#27](https://github.com/hop-top/kit/issues/27)) ([eda1458](https://github.com/hop-top/kit/commit/eda1458cd402cb81717abbb0de39426bbedcea0e))
* **domain:** generic domain framework ([#23](https://github.com/hop-top/kit/issues/23)) ([4c34b1c](https://github.com/hop-top/kit/commit/4c34b1cd9c5b020744df3cd392f046dd3fe4602e))
* **ext:** shared extensibility framework ([#18](https://github.com/hop-top/kit/issues/18)) ([477a0f9](https://github.com/hop-top/kit/commit/477a0f9cddaaa7e84659d514ca6dc5b8b614ec1b))
* **job:** durable job queue with multi-engine adapters ([#28](https://github.com/hop-top/kit/issues/28)) ([3e63582](https://github.com/hop-top/kit/commit/3e6358236a6423a84c8932a47a8c322239630256))
* **job:** install scripts, docker-compose, CI workflow ([#29](https://github.com/hop-top/kit/issues/29)) ([df78e4e](https://github.com/hop-top/kit/commit/df78e4ec0cb1787243daee5c4db0b8cfd7551749))
* **llm:** add missing LLM providers ([#15](https://github.com/hop-top/kit/issues/15)) ([b185f69](https://github.com/hop-top/kit/commit/b185f69214ac184dd8b630ec27042765b9fd3e44))
* **llm:** add provider-agnostic LLM package with OpenAI, Anthropic, and Ollama adapters ([4f5c646](https://github.com/hop-top/kit/commit/4f5c6460c0db869dc04030b5caca1d3948523a5d))
* **llm:** add TypeScript + Python ports and documentation ([3bd6281](https://github.com/hop-top/kit/commit/3bd6281c9da2743bc569280e8cf5ab96b10bc9ef))
* **llm:** integrate routellm + eva for cost-quality routing ([#9](https://github.com/hop-top/kit/issues/9)) ([e1c02d2](https://github.com/hop-top/kit/commit/e1c02d24532a461949105567506f40d427edd45f))
* **llm:** multimodal support — image, audio, generation ([#8](https://github.com/hop-top/kit/issues/8)) ([fa9725a](https://github.com/hop-top/kit/commit/fa9725ace4030e9c9ff8d4f09578579bdcaaf51e))
* **llm:** provider-agnostic LLM package ([a8eca8c](https://github.com/hop-top/kit/commit/a8eca8cd920cd7114ddbd550a870599307849b35))
* **llm:** TypeScript + Python ports and documentation ([1b0958d](https://github.com/hop-top/kit/commit/1b0958dae72d90cfada0a8db636d7032490c997e))
* **log:** add viper-configured charm.land/log/v2 wrapper ([9259cca](https://github.com/hop-top/kit/commit/9259cca28ebf1144e435531b4e850b4b86b19d5c))
* **markdown:** add glamour v2 rendering wrapper ([aacef1a](https://github.com/hop-top/kit/commit/aacef1a6d58382d6c821255ba459c8168d206371))
* **output:** add RegisterFlags to own --format flag ([c6a90e2](https://github.com/hop-top/kit/commit/c6a90e2895f6a6e20566c56442d046948d6c2a78))
* **output:** add table/json/yaml renderer ([14ecc29](https://github.com/hop-top/kit/commit/14ecc2934a992f621f17e9828d540bd91b684618))
* **output:** progressive disclosure hint subsystem ([bc3c1b5](https://github.com/hop-top/kit/commit/bc3c1b54367b0d3c98a91a063e288156307f5ead))
* **output:** progressive disclosure hint subsystem ([141b135](https://github.com/hop-top/kit/commit/141b135fe9fbe613958f6d7971cd50ec3de254b1))
* **ps:** cross-tool status convention package ([#33](https://github.com/hop-top/kit/issues/33)) ([5cae47a](https://github.com/hop-top/kit/commit/5cae47a828e3457458c72643526f786408ed5b1f))
* **release:** promotion system — scripts, gates, CODEOWNERS, docs ([65d83e6](https://github.com/hop-top/kit/commit/65d83e6f56de992585f2d602fdd9884fe5dcdef5))
* **router:** polyglot routing framework (Go + TS) ([#10](https://github.com/hop-top/kit/issues/10)) ([676ef40](https://github.com/hop-top/kit/commit/676ef405fd4505d0dc629e43100a660e1135f0fe))
* **sqlstore:** add BackupBeforeMigrate for pre-migration DB snapshots ([#11](https://github.com/hop-top/kit/issues/11)) ([892e0d6](https://github.com/hop-top/kit/commit/892e0d6b3d0073a159c63f050f3d1d300be211f9))
* **sqlstore:** add generic SQLite kv store with TTL ([285a10e](https://github.com/hop-top/kit/commit/285a10e534c4ab84d6689e81aec3b07d065f1f18))
* **templates:** CLI app template generation ([#14](https://github.com/hop-top/kit/issues/14)) ([36e7696](https://github.com/hop-top/kit/commit/36e76961e570037b5827de34666f707421a8a77e))
* **toolspec/usp:** learn workflows from real sessions ([#35](https://github.com/hop-top/kit/issues/35)) ([ab3dc3a](https://github.com/hop-top/kit/commit/ab3dc3a400a0f463c591062c5ec22734e89f7b20))
* **toolspec:** 12-factor AI CLI alignment (re-land) ([#40](https://github.com/hop-top/kit/issues/40)) ([79541ca](https://github.com/hop-top/kit/commit/79541caa20026d92422eab1faf74a0983bd8793f))
* **toolspec:** shared tool knowledge base ([#16](https://github.com/hop-top/kit/issues/16)) ([14410aa](https://github.com/hop-top/kit/commit/14410aa25456186fcb9dd02b84afa2c2e72a7e99))
* **tui:** add pre-themed bubbletea/v2 components ([c4a6610](https://github.com/hop-top/kit/commit/c4a66105c3ef862e966f755bdddb94482216692f))
* **tui:** add themed factory methods for table, textinput, and list ([19bc860](https://github.com/hop-top/kit/commit/19bc86044ba8698f2de65895826adfb69fe5d028))
* **tui:** TUI components architecture ([#13](https://github.com/hop-top/kit/issues/13)) ([5741dce](https://github.com/hop-top/kit/commit/5741dce26c2cd4f153bfb8704ed79eb860568355))
* **upgrade:** re-export hop.top/upgrade ([a6dc09d](https://github.com/hop-top/kit/commit/a6dc09d4a6019a28aee6fbad42b64061cc1a6106))
* **upgrade:** versioned schema migration with backup + rollback ([#32](https://github.com/hop-top/kit/issues/32)) ([00c1a0d](https://github.com/hop-top/kit/commit/00c1a0d7c68643147b5c3f65d7d7ad40ce5359d8))
* **uxp:** shared AI CLI detection & interfaces ([#17](https://github.com/hop-top/kit/issues/17)) ([46e9aef](https://github.com/hop-top/kit/commit/46e9aefd8184120e757d3bb69a62f106b17f2ac7))
* **wizard:** reusable onboarding wizard framework ([#38](https://github.com/hop-top/kit/issues/38)) ([669c80a](https://github.com/hop-top/kit/commit/669c80a7cfd9ca1bb8712f0da0dd30b684368829))
* **xdg:** add XDG path resolution ([db76e39](https://github.com/hop-top/kit/commit/db76e39f7f04e752bfbeb1aff18e5657de50a482))

### Bug Fixes

* address PR review feedback ([7fc3f0a](https://github.com/hop-top/kit/commit/7fc3f0aec3b4ed99d9790c25c356bc0a401a85b8))
* address PR review feedback ([f81ce42](https://github.com/hop-top/kit/commit/f81ce4282152f9076c799c375e8c7b47c3b7b67e))
* address PR review feedback ([fcf786b](https://github.com/hop-top/kit/commit/fcf786bc9b6379c05a4a0767477789b6f5154673))
* **api:** use stable messages in MapError, not domain error strings ([b0c95ce](https://github.com/hop-top/kit/commit/b0c95ce67da7e787d34fbed411a30db5899056bf))
* **cli:** resolve pino esbuild bundle failure in parity tests ([f771148](https://github.com/hop-top/kit/commit/f771148dbcff417703e7ccd126f1ac014e5d96ad))
* **kitadapter:** replace aspirational TODO with current limitation doc ([be9baae](https://github.com/hop-top/kit/commit/be9baae5ce455a0d72f56b9db582a4008ffdf6a9))
* **llm:** address PR review feedback ([48b5320](https://github.com/hop-top/kit/commit/48b5320b60b327fdba0f4f9b24749cdb92ba96fe))
* **markdown:** use hop.top accent for headings and links ([a810634](https://github.com/hop-top/kit/commit/a81063479a2a352d9bd3e946090f209ffa3ec549))
* **router:** remove unimplemented check, evaluate, logs command stubs ([dbcd335](https://github.com/hop-top/kit/commit/dbcd335e93d8194769ce0d5ee753a20f2cb2068b))
* **toolspec:** FindCommand uses BFS to prioritize shallow matches ([#36](https://github.com/hop-top/kit/issues/36)) ([ec67b46](https://github.com/hop-top/kit/commit/ec67b461f85680f51252c7b271acdaa3fc50da7b))
