# Changelog

## [0.10.0-beta](https://github.com/instill-ai/connector/compare/v0.9.0-beta...v0.10.0-beta) (2024-01-15)


### Features

* add end-user messages to HTTP errors ([#92](https://github.com/instill-ai/connector/issues/92)) ([d597648](https://github.com/instill-ai/connector/commit/d597648972b4eda8216948f24eaa4b09f490c7df))
* Extend Pinecone tasks with namespace and threshold ([#106](https://github.com/instill-ai/connector/issues/106)) ([17c1fc8](https://github.com/instill-ai/connector/commit/17c1fc86cdab8b76fb973fb0c1e3e02d26908a7c))
* **instill:** adopt latest Instill Model task format ([#95](https://github.com/instill-ai/connector/issues/95)) ([84778a7](https://github.com/instill-ai/connector/commit/84778a7c209383b897cd73fe3a4d67354dd04eb9))
* **instill:** enforces chat_history order pattern ([#99](https://github.com/instill-ai/connector/issues/99)) ([9bc4048](https://github.com/instill-ai/connector/commit/9bc404800d3e60abcf0f7786146ad26014bb0d49))
* **instill:** generate enumeration for `model_name` automatically ([#100](https://github.com/instill-ai/connector/issues/100)) ([dabfc19](https://github.com/instill-ai/connector/commit/dabfc199fa3f45f8ddb33d58146bf756a656ba61))
* **instill:** mark `prompt_images` as required in TASK_VISUAL_QUESTION_ANSWERING ([#110](https://github.com/instill-ai/connector/issues/110)) ([b0c21bf](https://github.com/instill-ai/connector/commit/b0c21bf89b9c7188368152726bfa4d0b0f7fdb75))
* **instill:** unify the chat_history format across different LLM tasks ([#96](https://github.com/instill-ai/connector/issues/96)) ([b09e7dd](https://github.com/instill-ai/connector/commit/b09e7ddab1d176006f27157ca6c9ec552fdb36fe))
* **instill:** use grpc client for all request ([#108](https://github.com/instill-ai/connector/issues/108)) ([548a78d](https://github.com/instill-ai/connector/commit/548a78d9ac89312c2060f3a8fc285a991cd9dd5d))
* **restapi:** inject the `output_body_schema` into component OpenAPI schema ([#101](https://github.com/instill-ai/connector/issues/101)) ([bd68d14](https://github.com/instill-ai/connector/commit/bd68d14da5ce613a66d9a75da0f52594abb73c75))
* **restapi:** remove base_url in restapi connector ([#102](https://github.com/instill-ai/connector/issues/102)) ([34d1a20](https://github.com/instill-ai/connector/commit/34d1a20784b4069fbdf7681622426d3f5db57a07))
* **website:** add `https` protocol to the url automatically ([#97](https://github.com/instill-ai/connector/issues/97)) ([62eb7e2](https://github.com/instill-ai/connector/commit/62eb7e216c20f95df111d3b69e58685286e05729))


### Bug Fixes

* **googlesearch,website:** fix potential memory leak by disabling http keep-alive ([#103](https://github.com/instill-ai/connector/issues/103)) ([7613561](https://github.com/instill-ai/connector/commit/76135615c56ec1b6d554cc1102b27d2f16a066bd))
* **instill:** accumulate the pagination when getting models ([#112](https://github.com/instill-ai/connector/issues/112)) ([cd077b0](https://github.com/instill-ai/connector/commit/cd077b00fe0df8c27f1ea7ef9a33f9b5b8c92801))
* **instill:** fix wrong required field in json-schema ([#98](https://github.com/instill-ai/connector/issues/98)) ([2d04474](https://github.com/instill-ai/connector/commit/2d0447433b17e9ab25e6a273398c18c72138259f))
* **stabilityai:** add the missing datauri prefix in image-to-image task ([#105](https://github.com/instill-ai/connector/issues/105)) ([e89f7ec](https://github.com/instill-ai/connector/commit/e89f7ec3b5424eb30a00fd8e283437f4f056bc45))

## [0.9.0-beta](https://github.com/instill-ai/connector/compare/v0.8.1-beta...v0.9.0-beta) (2024-01-01)


### Features

* **airbyte:** wrap all Airbyte connectors into one ([#79](https://github.com/instill-ai/connector/issues/79)) ([30fe290](https://github.com/instill-ai/connector/commit/30fe2900bd9a74273e235ab5f6ab60b10e3376c3))
* **numbers:** migrate to Capture API ([#89](https://github.com/instill-ai/connector/issues/89)) ([e976854](https://github.com/instill-ai/connector/commit/e9768548d817ecc77cdfa5572e7663fb55fbfa7b))
* support metadata in Pinecone connector ([#87](https://github.com/instill-ai/connector/issues/87)) ([3734773](https://github.com/instill-ai/connector/commit/37347730cdd8dc25c34b7753cf8b6eb653b9e327))


### Bug Fixes

* **instill:** fix wrong Airbyte image_name ([#91](https://github.com/instill-ai/connector/issues/91)) ([52e8409](https://github.com/instill-ai/connector/commit/52e8409522437e0064627ddc7067a07615e9fe5f))

## [0.8.1-beta](https://github.com/instill-ai/connector/compare/v0.8.0-beta...v0.8.1-beta) (2023-12-22)


### Features

* Improve error messages in connector execution [#311](https://github.com/instill-ai/connector/issues/311)  ([#76](https://github.com/instill-ai/connector/issues/76)) ([d0dea69](https://github.com/instill-ai/connector/commit/d0dea69b3d0ccbfdfbcdef54a1e8fdbbefe264e4))


### Miscellaneous Chores

* release v0.8.1-beta ([692c72a](https://github.com/instill-ai/connector/commit/692c72a8070e3b97601e57bb414080c9ce9ad9b3))

## [0.8.0-beta](https://github.com/instill-ai/connector/compare/v0.7.0-alpha...v0.8.0-beta) (2023-12-15)


### Features

* **instill:** add new tasks ([#81](https://github.com/instill-ai/connector/issues/81)) ([c0a3725](https://github.com/instill-ai/connector/commit/c0a3725a8bfc5d0ed1cf063fa502ba9c1b8c869a))


### Bug Fixes

* **instill:** add mime prefix to image output ([959a69d](https://github.com/instill-ai/connector/commit/959a69d167d556792555ac1400198b4a3117748f))
* **redis:** fix message retrieval and improve system message support ([#83](https://github.com/instill-ai/connector/issues/83)) ([0c19492](https://github.com/instill-ai/connector/commit/0c19492e7fa67335a32ae556140b8db8c6ffd047))
* **website:** correct the field json mappings ([#72](https://github.com/instill-ai/connector/issues/72)) ([ed45f6f](https://github.com/instill-ai/connector/commit/ed45f6f6169f4117e564c79685ed828938f542d9))


### Miscellaneous Chores

* release v0.8.0-beta ([0548a63](https://github.com/instill-ai/connector/commit/0548a63ebc4d8c65322e121774346b8e126f9f67))

## [0.7.0-alpha](https://github.com/instill-ai/connector/compare/v0.6.0-alpha...v0.7.0-alpha) (2023-11-28)


### Features

* **openai:** support text to speech task ([#52](https://github.com/instill-ai/connector/issues/52)) ([7c3caf7](https://github.com/instill-ai/connector/commit/7c3caf76db144e6ee074f4d5f106b905fc3f68b5))
* **redis,openai:** support redis as LLM chat memory store ([#53](https://github.com/instill-ai/connector/issues/53)) ([bf5dea7](https://github.com/instill-ai/connector/commit/bf5dea7cf81d09637b638b3b11be003c08dd2da1))
* **redis:** add SSL/TLS support for Redis ([#62](https://github.com/instill-ai/connector/issues/62)) ([450b60d](https://github.com/instill-ai/connector/commit/450b60d30c7cd376cbd7e5ef81d6bedf278faf56))
* **restapi:** add REST API connector ([#54](https://github.com/instill-ai/connector/issues/54)) ([a795462](https://github.com/instill-ai/connector/commit/a795462922c7525d46ba3ae509447b29a8733226))
* **website,googlesearch:** add website connector and improve webpage text parsing ([#64](https://github.com/instill-ai/connector/issues/64)) ([879904f](https://github.com/instill-ai/connector/commit/879904f5e39d60cf487e1cea078a32b617042bc6))


### Bug Fixes

* **stability-ai:** add default weight for stable-diffusion-xl-1024-v1-0 ([#61](https://github.com/instill-ai/connector/issues/61)) ([7c18737](https://github.com/instill-ai/connector/commit/7c18737c606bb62e3a2b0bd3ee7e2d00047c849d))

## [0.6.0-alpha](https://github.com/instill-ai/connector/compare/v0.5.0-alpha...v0.6.0-alpha) (2023-11-11)


### Features

* **ai-openai:** support OpenAI gpt-4-turbo and dall-e-3 ([#43](https://github.com/instill-ai/connector/issues/43)) ([38c451e](https://github.com/instill-ai/connector/commit/38c451e532764a9f1ec5c25abe0d87f5078dcde1))
* **google-search:** support google search connector ([#41](https://github.com/instill-ai/connector/issues/41)) ([950510e](https://github.com/instill-ai/connector/commit/950510ea5a2bddbfa1d1ad9af8393ddd7bbca680))


### Bug Fixes

* **google-search:** fix google search nil pointer ([#48](https://github.com/instill-ai/connector/issues/48)) ([d681159](https://github.com/instill-ai/connector/commit/d6811595a1e79493f6cf78a1b40594b43523f4fd))
* **openai:** fix the message order in the chat completion request ([#36](https://github.com/instill-ai/connector/issues/36)) ([7ef3177](https://github.com/instill-ai/connector/commit/7ef3177c74149cc7818916dc4c81e90d0dbd84d3))

## [0.5.0-alpha](https://github.com/instill-ai/connector/compare/v0.4.0-alpha...v0.5.0-alpha) (2023-10-27)


### Miscellaneous Chores

* **release:** release v0.5.0-alpha ([58883d9](https://github.com/instill-ai/connector/commit/58883d9b112a6057f60ba530d749103f191b517a))

## [0.4.0-alpha](https://github.com/instill-ai/connector/compare/v0.3.0-alpha...v0.4.0-alpha) (2023-09-13)


### Miscellaneous Chores

* **release:** release v0.4.0-alpha ([725b63f](https://github.com/instill-ai/connector/commit/725b63f948366db1670b2b0d34a0309c5ebb06c6))

## [0.3.0-alpha](https://github.com/instill-ai/connector/compare/v0.2.0-alpha...v0.3.0-alpha) (2023-08-03)


### Miscellaneous Chores

* **release:** release v0.3.0-alpha ([dfe81c0](https://github.com/instill-ai/connector/commit/dfe81c05fea87a887f94628d3908251961c0058e))

## [0.2.0-alpha](https://github.com/instill-ai/connector/compare/v0.1.0-alpha...v0.2.0-alpha) (2023-07-20)


### Miscellaneous Chores

* **release:** release v0.2.0-alpha ([fa946bd](https://github.com/instill-ai/connector/commit/fa946bd6ad4984ecba92e5fd9d0c477bd7fe21ef))

## [0.1.0-alpha](https://github.com/instill-ai/connector/compare/v0.1.0-alpha...v0.1.0-alpha) (2023-07-09)


### Features

* Added object mapper implementation and basic tests ([#7](https://github.com/instill-ai/connector/issues/7)) ([a91364b](https://github.com/instill-ai/connector/commit/a91364b7e08866259296810743803341a2097612))


### Miscellaneous Chores

* **release:** release v0.1.0-alpha ([6984052](https://github.com/instill-ai/connector/commit/6984052f8e5a80201b90e82580f10f8872c86d7e))
