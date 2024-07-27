import retvec
import datasets
import os

os.environ['TF_CPP_MIN_LOG_LEVEL'] = '1'  # silence TF INFO messages
import tensorflow as tf
import numpy as np
from tensorflow.keras import layers
from retvec.tf import RETVecTokenizer

NUM_CLASSES = 3


def getData(folder_path):
    labels = []
    msgs = []
    # 遍历文件夹
    for root, dirs, files in os.walk(folder_path):
        # 遍历当前文件夹下的所有文件
        for filename in files:
            # 判断是否为csv文件
            if filename.endswith(".csv"):
                file_path = os.path.join(root, filename)
                # 读取csv文件内容
                with open(file_path, 'r', errors='ignore') as csv_file:
                    for line in csv_file:
                        if line[0] == '' or line[0]==' ':
                            continue
                        labels.append([int(str.strip(line[0]))])
                        msgs.append(line[3:])
    return np.array(msgs), np.array(labels)


trainDataMsgs, trainDataLabels = getData("./trainData")
testDataMsgs, testDataLabels = getData("./testData")


# preparing data
x_train = tf.constant(trainDataMsgs, dtype=tf.string)

print(x_train.shape)

y_train = np.zeros((len(x_train),NUM_CLASSES))
for idx, ex in enumerate(trainDataLabels):
    for val in ex:
        y_train[idx][val] = 1


# test data
x_test = tf.constant(testDataMsgs, dtype=tf.string)
y_test = np.zeros((len(x_test),NUM_CLASSES))
for idx, ex in enumerate(testDataLabels):
    for val in ex:
        y_test[idx][val] = 1


# using strings directly requires to put a shape of (1,) and dtype tf.string
inputs = layers.Input(shape=(1,), name="token", dtype=tf.string)

# add RETVec tokenizer layer with default settings -- this is all you have to do to build a model with RETVec!
x = RETVecTokenizer(model='retvec-v1')(inputs)

# standard two layer LSTM
x = layers.Bidirectional(layers.LSTM(64, return_sequences=True))(x)
x = layers.Bidirectional(layers.LSTM(64))(x)
outputs = layers.Dense(NUM_CLASSES, activation='sigmoid')(x)
model = tf.keras.Model(inputs, outputs)
model.summary()

# compile and train the model
batch_size = 256
epochs = 2
model.compile('adam', 'binary_crossentropy', ['acc'])
history = model.fit(x_train, y_train, epochs=epochs, batch_size=batch_size,
                    validation_data=(x_test, y_test))

# saving the model
save_path = './emotion_model/1'
model.save(save_path)
