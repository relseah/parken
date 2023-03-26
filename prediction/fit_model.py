import csv
import pickle

from sklearn.model_selection import cross_val_score, cross_validate, train_test_split
from sklearn.ensemble import RandomForestRegressor
import numpy as np


def load_data(filename):
    with open(filename) as file:
        rows = [[int(datapoint) for datapoint in row]
                for row in csv.reader(file)]
        return rows[0] if len(rows) == 1 else rows


def save_model(model, filename):
    with open(filename, 'wb') as file:
        pickle.dump(model, file, protocol=pickle.HIGHEST_PROTOCOL)


def fit_model():
    data = load_data('data.csv')
    target = load_data('target.csv')
    X, y = np.array(data), np.array(target)
    model = RandomForestRegressor(random_state=0, n_jobs=-1)
    X_train, X_test, y_train, y_test = train_test_split(X, y, random_state=1)
    return model.fit(X_train, y_train)
    """
    print(model.score(X_train, y_train))
    print(model.score(X_test, y_test))
    print(cross_val_score(model, X, y, n_jobs=-1))
    result = cross_validate(
        model, X, y, scoring='neg_mean_squared_error', n_jobs=-1)
    pprint(result)
    """


if __name__ == '__main__':
    model = fit_model()
    save_model(model, 'model.dat')
