# Changelog

## [0.2.0](https://github.com/cloudnative-pg/machinery/compare/v0.1.0...v0.2.0) (2025-03-27)


### Features

* add pg_config interface ([#88](https://github.com/cloudnative-pg/machinery/issues/88)) ([0fc1fae](https://github.com/cloudnative-pg/machinery/commit/0fc1faed3332d667e2cdec42a90596e4644920a9))


### Bug Fixes

* **deps:** update all non-major go dependencies ([#84](https://github.com/cloudnative-pg/machinery/issues/84)) ([fdb8e45](https://github.com/cloudnative-pg/machinery/commit/fdb8e4548fde9ca033b46b8551593d13dbbfec3d))
* **deps:** update all non-major go dependencies ([#90](https://github.com/cloudnative-pg/machinery/issues/90)) ([588ff73](https://github.com/cloudnative-pg/machinery/commit/588ff73495e544c0da6f8872893de4230c3240ae))
* **deps:** update module k8s.io/apimachinery to v0.32.3 ([#86](https://github.com/cloudnative-pg/machinery/issues/86)) ([ad50665](https://github.com/cloudnative-pg/machinery/commit/ad506659a61762d911c7a879fd3682df8d5514ae))
* **deps:** update module sigs.k8s.io/controller-runtime to v0.20.4 ([#85](https://github.com/cloudnative-pg/machinery/issues/85)) ([0384357](https://github.com/cloudnative-pg/machinery/commit/0384357da045806d19fa3f05000f70fdb8d6ba1f))
* typo ([#92](https://github.com/cloudnative-pg/machinery/issues/92)) ([c1ff6ce](https://github.com/cloudnative-pg/machinery/commit/c1ff6ce96607dd11ed08022096e67516f85181a1))

## 0.1.0 (2025-02-26)


### Features

* add `GatherReadyWALFiles` ([#54](https://github.com/cloudnative-pg/machinery/issues/54)) ([2807bc8](https://github.com/cloudnative-pg/machinery/commit/2807bc88310dbfa0f0c284d175dc932b1d031da9))
* add envmap package ([#53](https://github.com/cloudnative-pg/machinery/issues/53)) ([257ab8d](https://github.com/cloudnative-pg/machinery/commit/257ab8d1e6a2d9a50710c97f03de4e88e284dee7))
* add functions mapping LSNs to WAL file names ([#65](https://github.com/cloudnative-pg/machinery/issues/65)) ([a3f9167](https://github.com/cloudnative-pg/machinery/commit/a3f9167a392a58f472d51c266f3576aa3c32776b))
* add stringset ([#33](https://github.com/cloudnative-pg/machinery/issues/33)) ([267a543](https://github.com/cloudnative-pg/machinery/commit/267a543ce26f1e61149d9880eb76d58c85524886))
* add stringset intersect ([#37](https://github.com/cloudnative-pg/machinery/issues/37)) ([5ac7af3](https://github.com/cloudnative-pg/machinery/commit/5ac7af31ef720a6196443bfa1976f786b01d1960))
* initial import ([#1](https://github.com/cloudnative-pg/machinery/issues/1)) ([2574dac](https://github.com/cloudnative-pg/machinery/commit/2574dac6e45ab5ab1e9132eb6b884335a8cb7b81))
* **lsn:** add `LSNStartFromWALName` ([#69](https://github.com/cloudnative-pg/machinery/issues/69)) ([77a23bc](https://github.com/cloudnative-pg/machinery/commit/77a23bcd05c3fce1e029e040ae703678e7e3b53c))
* nanosecond-precision logging ([#38](https://github.com/cloudnative-pg/machinery/issues/38)) ([a791d08](https://github.com/cloudnative-pg/machinery/commit/a791d08903bfd3f1b7becaa25b80cc9311f2d9b3))
* time-management functions ([#32](https://github.com/cloudnative-pg/machinery/issues/32)) ([1e197af](https://github.com/cloudnative-pg/machinery/commit/1e197af1f392697787e671db1ca65902c0b6ccbb))
* types/functions for images and Postgres versions ([#29](https://github.com/cloudnative-pg/machinery/issues/29)) ([34c8797](https://github.com/cloudnative-pg/machinery/commit/34c8797af80f3c980cb33674a6e00a44cb777825))


### Bug Fixes

* archive WAL file once, not twice ([#77](https://github.com/cloudnative-pg/machinery/issues/77)) ([ef857fb](https://github.com/cloudnative-pg/machinery/commit/ef857fb8ea8ef8dde17240debdc2e99b306588a3))
* **deps:** update all non-major go dependencies ([#45](https://github.com/cloudnative-pg/machinery/issues/45)) ([281395e](https://github.com/cloudnative-pg/machinery/commit/281395ea76dabbeb759e8e0c24efa33f4ed49513))
* **deps:** update all non-major go dependencies ([#49](https://github.com/cloudnative-pg/machinery/issues/49)) ([66cd032](https://github.com/cloudnative-pg/machinery/commit/66cd032ef6072aea5313b367591bbc3715f166bb))
* **deps:** update all non-major go dependencies ([#62](https://github.com/cloudnative-pg/machinery/issues/62)) ([95c37fe](https://github.com/cloudnative-pg/machinery/commit/95c37fe624d0035055d77dd0beb30333f3844507))
* **deps:** update all non-major go dependencies ([#72](https://github.com/cloudnative-pg/machinery/issues/72)) ([81a93a4](https://github.com/cloudnative-pg/machinery/commit/81a93a4d6ef82ac27e773c4b1171d4ed72739711))
* **deps:** update module golang.org/x/sys to v0.26.0 ([#34](https://github.com/cloudnative-pg/machinery/issues/34)) ([c27747f](https://github.com/cloudnative-pg/machinery/commit/c27747f9974b422b6b7cbe41c6c195ecfa8736d5))
* **deps:** update module golang.org/x/sys to v0.29.0 ([#63](https://github.com/cloudnative-pg/machinery/issues/63)) ([2553c23](https://github.com/cloudnative-pg/machinery/commit/2553c239f2c8af8adfa28d6b5820bb08e574d7ab))
* **deps:** update module k8s.io/apimachinery to v0.31.2 ([#41](https://github.com/cloudnative-pg/machinery/issues/41)) ([042a028](https://github.com/cloudnative-pg/machinery/commit/042a028b767c0ea741995c9c1d9149caab800061))
* **deps:** update module k8s.io/apimachinery to v0.32.2 ([#55](https://github.com/cloudnative-pg/machinery/issues/55)) ([ff7f7bf](https://github.com/cloudnative-pg/machinery/commit/ff7f7bf5f301808a6e7de6aeaa1aaf35fadfb2bc))
* **deps:** update module sigs.k8s.io/controller-runtime to v0.19.1 ([#19](https://github.com/cloudnative-pg/machinery/issues/19)) ([fa7f9a9](https://github.com/cloudnative-pg/machinery/commit/fa7f9a984af46c9dae796f71310728b56085df88))
* **deps:** update module sigs.k8s.io/controller-runtime to v0.20.2 ([#56](https://github.com/cloudnative-pg/machinery/issues/56)) ([015b210](https://github.com/cloudnative-pg/machinery/commit/015b21025c63a91080b5c92d6999464e74a39165))
* use uint64 for WAL offset and sizes ([#66](https://github.com/cloudnative-pg/machinery/issues/66)) ([ac9ba00](https://github.com/cloudnative-pg/machinery/commit/ac9ba00698fcfe538c3115c340b0f7b9b652f05e))
