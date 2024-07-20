import os
import numpy as np
os.environ['TF_CPP_MIN_LOG_LEVEL'] = '1'  # silence TF INFO messages
import tensorflow as tf

save_path = './emotion_model/1'

model = tf.keras.models.load_model(save_path, compile=False)

model.summary()

CLASSES = {
    0:'普通邮件',
    1:'广告邮件',
    2:'诈骗邮件'
}

def predict_emotions(txt):
    # recall it is multi-class so we need to get all prediction above a threshold (0.5)
    input = tf.constant( np.array([txt]) , dtype=tf.string )

    preds = model(input)[0]
    maxClass = -1
    maxScore = 0
    for idx in range(3):
        if preds[idx] > maxScore:
            maxScore = preds[idx]
            maxClass = idx
    return maxClass


maxClass = predict_emotions("各位同事请注意 这里是110，请大家立刻把银行卡账号密码回复发给我！")

print("这个邮件属于：",CLASSES[maxClass])