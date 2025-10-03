# vdsymlink

docker镜像地址：https://hub.docker.com/r/susy803/vdsymlink

对视频批量建立软链接和格式化命名

源目录请设置为正片所在目录

以下为示例：（注意飞牛OS暂时不支持符号链接，所以生成的符号链接无法直接在飞牛内双击播放，windows通过samba挂载可以）

```bash
/vol1/1000/download/[Snow-Raws] ラグナクリムゾン
├── 映像特典
│   ├──[Snow-Raws] ラグナクリムゾン ｢感脳性リベレーション｣ ノンクレジットバージョン 13～21話 (BD 1920x1080 HEVC-YUV420P10 FLAC).mkv
│   ├──[Snow-Raws] ラグナクリムゾン Blu-ray BOX Ⅱ MENU_00 (BD 1920x1080 HEVC-YUV420P10 FLAC).mkv
│   └──[Snow-Raws] ラグナクリムゾン 放送直前!! 緊急配信番組 ～狩竜人の夜明け～（ディレクターズカット版） (BD 1920x1080 HEVC-YUV420P10 FLAC).mkv
├── 特典CD
│   └──ラグナクリムゾン オリジナルサウンドトラック Vol.1'  'ラグナクリムゾン オリジナルサウンドトラック
│       └── 01 曲目 1.flac
├── [Snow-Raws] ラグナクリムゾン 第01話 (BD 1920x1080 HEVC-YUV420P10 FLAC).mkv
├── [Snow-Raws] ラグナクリムゾン 第02話 (BD 1920x1080 HEVC-YUV420P10 FLAC).mkv
├── ...
├── [Snow-Raws] ラグナクリムゾン 第23話 (BD 1920x1080 HEVC-YUV420P10 FLAC).mkv
└── [Snow-Raws] ラグナクリムゾン 第24話 (BD 1920x1080 HEVC-YUV420P10 FLAC).mkv

目标路径：/media/狩龙人拉格纳
生成符号链接
/media/狩龙人拉格纳
└── [Snow-Raws] ラグナクリムゾン
    └── S01
        ├── [Snow-Raws] ラグナクリムゾン.S01E01.mkv
        ├── [Snow-Raws] ラグナクリムゾン.S01E02.mkv
        ├── ...
        ├── [Snow-Raws] ラグナクリムゾン.S01E23.mkv
        └── [Snow-Raws] ラグナクリムゾン.S01E24.mkv

目标路径：/media/狩龙人拉格纳/S01
生成符号链接
/media/狩龙人拉格纳
└── S01
    ├── 狩龙人拉格纳.S01E01.mkv
    ├── 狩龙人拉格纳.S01E02.mkv
    ├── ...
    ├── 狩龙人拉格纳.S01E23.mkv
    └── 狩龙人拉格纳.S01E24.mkv
```

如果使用emby/jellfin为docker版本，请映射源文件地址到容器内部，请注意：

``` bash
如果不填写重定向，请保证映射路径要从宿主机根路径填写完整
/vol1/1000/download/:/vol1/1000/download

如果填写重定向路径则随意，例如
/vol1/1000/download/:/download
但是请填写重定向路径填写为容器内部映射路径，这里为/download

重定向后的符号链接无法直接在宿主机上播放，请通过emby/jellyfin播放
```
