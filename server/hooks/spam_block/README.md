# 插件介绍

使用机器学习的方式识别垃圾邮件。模型使用的是RETVec。模型参数约 200k，在我1核1G的服务器上，单次推理耗时约3秒，Mac M1上可达到毫秒级耗时。耗时上，其实可以将模型进行裁剪，转换为Tensorflow Lite模型，转换后模型的资源消耗应该更小。但是Lite模型部署比较繁琐，涉及大量C库的编译安装，过程过于复杂。另外
我觉得，这个模型在我这垃圾服务器上面都能勉强使用，其他所有人的服务器上面应该都能顺利运行了，没必要继续裁剪模型了。

# Help

目前Google GMail使用的垃圾邮件识别算法也是RETVec，理论上识别效果能够达到GMail同等级别。但是，我并没有Google那样大量的训练集。欢迎大家以Pull
Request的形式提交机器学习的各类样本数据。

你可以给testData和trainData这两个文件夹下面的csv文件提交PR，CSV文件每行的第一个数字表示数据类型，0表示正常邮件，1表示广告邮件，2表示诈骗邮件。

你可以使用export.go这个脚本或者从Release中下载[email_export工具](https://github.com/Jinnrry/PMail/releases/tag/v2.6.1)导出你全部的邮件数据，过滤掉隐私内容并且标记好分类后提交上来。

# 如何运行

1、下载[emotion_model.zip](https://github.com/Jinnrry/PMail/releases/tag/v2.6.1)或者自己训练模型

2、使用docker运行tensorflow模型
`docker run -d -p 127.0.0.1:8501:8501   \
-v "{模型文件位置}:/models/emotion_model"   \
-e MODEL_NAME=emotion_model     tensorflow/serving &`

3、CURL测试模型部署是否成功

> 详细部署说明请参考[tensorflow官方](https://www.tensorflow.org/tfx/guide/serving?hl=zh-cn)

```bash
curl -X POST http://localhost:8501/v1/models/emotion_model:predict -d '{ 
    "instances": [
        {"token":["各位同事请注意 这里是110，请大家立刻把银行卡账号密码回复发给我！"]}
    ]
}' 
```

将得到类似输出：

```json
{
  "predictions": [
    [
      0.394376636,
      // 正常邮件的得分
      0.0055413493,
      // 广告邮件的得分
      0.633584619
      // 诈骗邮件的得分，这里诈骗邮件得分最高，因此最可能为诈骗邮件
    ]
  ]
}
```

4、将spam_block插件移动到pmail插件目录

5、在插件位置新建配置文件`spam_block_config.json`内容类似

```json
{
  "apiURL": "http://localhost:8501/v1/models/emotion_model:predict",
  "apiTimeout": 3000
}
```

apiURL表示模型api访问地址，如果你是使用Docker部署，PMail和tensorflow/serving容器需要设置为相同网络才能通信，并且需要把localhost替换为tensorflow/serving的容器名称

# 模型效果

trec06c数据集：

loss: 0.0187 - acc: 0.9948 - val_loss: 0.0047 - val_acc: 0.9993

实际使用效果：

我最近一周的使用效果来看，实际使用效果远低于模型理论效果。猜测原因如下：

trec06c数据集已经公开十多年了，目前应该市面上所有反垃圾系统都使用这个数据集训练过。这个训练集训练出来的特征可能具有普遍性，而对于发垃圾邮件的人来说，这十多年他们也大致摸透了哪些特征会被识别为垃圾邮件，因此他们会针对性的避开很多关键字以免被封

解决方案只能是加入更多更优质的训练数据，但是trec06c之后就没这样优质的训练数据了，因此如果大家愿意，欢迎贡献模型训练数据。另外，针对模型本身，也欢迎提出优化方案。

# 训练模型

`python train.py`

# 测试模型

`python test.py`

# trec06c 数据集

[trec06c_format.py](trec06c_format.py)
脚本用于整理trec06c数据集，将其转化为训练所需的数据格式。由于数据集版权限制，如有需要请前往[这里](https://plg.uwaterloo.ca/~gvcormac/treccorpus06/about.html)
自行下载，本项目中不直接引入数据集内容。

# 致谢

Tanks For [google-research/retvec](https://github.com/google-research/retvec)

