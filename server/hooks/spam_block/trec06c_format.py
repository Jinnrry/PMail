import os
from email.parser import Parser
from email.policy import default

# 该脚本用于整理trec06c数据集，可以生成训练集和测试集数据格式

def getData(path):
    f = open(path, 'r', encoding='gb2312', errors='ignore')
    data = f.read()
    headers = Parser(policy=default).parsestr(data)
    body = headers.get_payload()
    body = body.replace("\n", "")

    return headers["subject"], body


num = 0

# getData("../data/000/000")
with open("index", "r") as f:
    with open("trec06c_train.csv", "w") as w:
        with open("trec06c_test.csv", "w") as wt:
            while True:
                line = f.readline()
                if not line:
                    break
                infos = line.split(" ")
                subject, body = getData(infos[1].strip())
                tp = 0
                if infos[0].lower() == "spam":
                    tp = 1
                data = "{} \t{} {}\n".format(tp, subject, body)
                if num < 55000:
                    w.write(data)
                else:
                    wt.write(data)
                num += 1
print(num)