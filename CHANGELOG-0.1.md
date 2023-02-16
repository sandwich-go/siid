### v0.1.0 🌈 (2023-02-16 19:16:11)

#### 🐛  Bug Fixed
  * 当发生id run out error的时候，会再次进行renew ([3a3d3f2](https://github.com/sandwich-go/siid/commit/3a3d3f2b468251a3878840877d38585d5b7e2800)) (<small>[祝黄清](huangqing.zhu@centurygame.com)@2023-02-16 19:16:11 &#43;0800 &#43;0800</small>)
  * fix test error ([fd8663c](https://github.com/sandwich-go/siid/commit/fd8663c7f68f78c8b2aed7d9dc975727b6ddf083)) (<small>[祝黄清](huangqing.zhu@centurygame.com)@2022-11-23 12:10:59 &#43;0800 &#43;0800</small>)

#### 🚀  New Feature
  * add builder.Range interface, range all engine ([eacc767](https://github.com/sandwich-go/siid/commit/eacc767fbb6731a8a2d67f8c5fd21e5e7efc9d36)) (<small>[祝黄清](huangqing.zhu@centurygame.com)@2022-11-25 16:29:28 &#43;0800 &#43;0800</small>)
  * 增加BuildWithOffset接口，用来build指定offsetOnCreate的Engine ([b39b412](https://github.com/sandwich-go/siid/commit/b39b412092a9b401515fb22d318440d4759756a2)) (<small>[祝黄清](huangqing.zhu@centurygame.com)@2022-11-24 15:29:34 &#43;0800 &#43;0800</small>)
  * 修改siid.New参数，传入*Options ([2af1ff7](https://github.com/sandwich-go/siid/commit/2af1ff7c1cc531f96709a172b39feb45f5df8fea)) (<small>[祝黄清](huangqing.zhu@centurygame.com)@2022-11-22 13:57:01 &#43;0800 &#43;0800</small>)
  * use sandwich-go/logbus ([f624fa6](https://github.com/sandwich-go/siid/commit/f624fa68e592f4fbde2c55b4737281c82641e33a)) (<small>[祝黄清](huangqing.zhu@centurygame.com)@2022-11-22 11:57:24 &#43;0800 &#43;0800</small>)
  * add renew_retry_delay parameter, remove retry func from options ([442e0c6](https://github.com/sandwich-go/siid/commit/442e0c6118908b8b71c8a4c3a42186fd27b8f979)) (<small>[祝黄清](huangqing.zhu@centurygame.com)@2022-11-22 11:09:03 &#43;0800 &#43;0800</small>)

#### 🛠  Refactor
  * should not lost error when panic ([ba83d0f](https://github.com/sandwich-go/siid/commit/ba83d0ff58a9d1eb535e712e212912dda5357f88)) (<small>[hui.wang](hui.wang@centurygame.com)@2023-02-16 18:24:31 &#43;0800 &#43;0800</small>)
  * 移动bench至bench module中 ([af6cb6a](https://github.com/sandwich-go/siid/commit/af6cb6a5fc8d21de84153819347153be56ef85fb)) (<small>[祝黄清](huangqing.zhu@centurygame.com)@2022-11-25 14:43:38 &#43;0800 &#43;0800</small>)

#### 🧪  Testing
  * 增加benchmark ([adea15d](https://github.com/sandwich-go/siid/commit/adea15d19146b803c770752cbd5fa91d1f60d03a)) (<small>[祝黄清](huangqing.zhu@centurygame.com)@2022-11-15 17:16:03 &#43;0800 &#43;0800</small>)

#### 🤖  Tools
  * 升级logbus ([65c1716](https://github.com/sandwich-go/siid/commit/65c1716374bfcfba719bbae83bf04566fd9e3c53)) (<small>[祝黄清](huangqing.zhu@centurygame.com)@2022-11-24 17:16:58 &#43;0800 &#43;0800</small>)
  * readme ([1c12497](https://github.com/sandwich-go/siid/commit/1c124972b09afe09b59bd5aaaf4818d243db8ddc) , [1314518](https://github.com/sandwich-go/siid/commit/1314518e8cca43b39eab43896efa9539ce708b8d)) (<small>[祝黄清](huangqing.zhu@centurygame.com)@2022-11-17 11:35:08 &#43;0800 &#43;0800</small>)
  * 性能报告 ([f36fb70](https://github.com/sandwich-go/siid/commit/f36fb702c8db9b580d5d9b90a55931a8aca21460)) (<small>[祝黄清](huangqing.zhu@centurygame.com)@2022-11-14 18:36:44 &#43;0800 &#43;0800</small>)
  * 整理reporter ([3923a69](https://github.com/sandwich-go/siid/commit/3923a691a77b617709091e2d26e69edfc37c89ec)) (<small>[祝黄清](huangqing.zhu@centurygame.com)@2022-11-14 10:59:00 &#43;0800 &#43;0800</small>)
  * 基本框架 ([94e9bfe](https://github.com/sandwich-go/siid/commit/94e9bfe7e2e443ef8690b93961fb8950940fa5ed)) (<small>[祝黄清](huangqing.zhu@centurygame.com)@2022-11-11 21:05:03 &#43;0800 &#43;0800</small>)

#### 💪  Commit
  * siid初始化 ([981790d](https://github.com/sandwich-go/siid/commit/981790dedb671ece2df8b5be2b254fde8c2252c7) , [c2f82f5](https://github.com/sandwich-go/siid/commit/c2f82f59cff3732a01a22872d75109598c6f0fe9) , [e855ab2](https://github.com/sandwich-go/siid/commit/e855ab205f711e1335415ce899851517d2eafb39) , [601b1ce](https://github.com/sandwich-go/siid/commit/601b1cef4f554c61c8f9a48acda965f5e951217e)) (<small>[祝黄清](huangqing.zhu@centurygame.com)@2022-11-14 18:00:35 &#43;0800 &#43;0800</small>)



