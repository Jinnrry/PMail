import os
from email.parser import Parser
from email.policy import default
from bs4 import BeautifulSoup


# 该脚本用于整理trec06c数据集，可以生成训练集和测试集数据格式

def getData(path):
    f = open(path, 'r', errors='ignore')
    data = f.read()
    headers = Parser(policy=default).parsestr(data)
    body = ""
    if headers.is_multipart():
        for part in headers.iter_parts():
            tbody = part.get_payload()
            if isinstance(tbody, list):
                for item in tbody:
                    txt = item.get_payload()
                    if isinstance(tbody, list):
                        return "", ""
                    bsObj = BeautifulSoup(txt, 'lxml')
                    body += bsObj.get_text()
            else:
                bsObj = BeautifulSoup(tbody, 'lxml')
                body += bsObj.get_text()
    else:
        tbody = headers.get_payload()
        bsObj = BeautifulSoup(tbody, 'lxml')
        body += bsObj.get_text()
    return headers["subject"], body.replace("\n", "")


num = 0

# getData("../data/000/000")
with open("index", "r") as f:
    with open("trec07p_train.csv", "w") as w:
        with open("trec07p_test.csv", "w") as wt:
            while True:
                line = f.readline()
                if not line:
                    break
                infos = line.split(" ")
                subject, body = getData(infos[1].strip())
                if subject == "":
                    continue
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
