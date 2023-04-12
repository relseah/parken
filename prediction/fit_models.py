import csv
import pickle

from sklearn.model_selection import cross_val_score, cross_validate, train_test_split
from sklearn.ensemble import RandomForestRegressor
import numpy as np

from prediction import mask_samples


def load_data(filename):
    with open(filename) as file:
        rows = [[int(datapoint) for datapoint in row]
                for row in csv.reader(file)]
        return rows[0] if len(rows) == 1 else rows


def save_model(model, filename):
    with open(filename, 'wb') as file:
        pickle.dump(model, file, protocol=pickle.HIGHEST_PROTOCOL)


def fit_model(X, y):
    model = RandomForestRegressor(random_state=0, n_jobs=-1)
    return model.fit(X, y)


def build_model(X, y):
    X_train, X_test, y_train, y_test = train_test_split(
        X, y, random_state=1)
    model = fit_model(X_train, y_train)
    print('Train score:', model.score(X_train, y_train))
    print('Test score:',  model.score(X_test, y_test))
    return model


if __name__ == '__main__':
    data_lt = load_data('data-long-term.csv')
    data_st = load_data('data-short-term.csv')
    target = load_data('target.csv')
    X_lt, y = np.array(data_lt), np.array(target)
    print('Long-term model')
    save_model(build_model(X_lt, y), 'model-long-term.dat')
    X_st = mask_samples(data_st)
    print(X_st.mask[0:10])
    print('Short-term model')
    save_model(build_model(X_st, y), 'model-short-term.dat')
    """
    print(cross_val_score(model, X, y, n_jobs=-1))
    result = cross_validate(
        model, X, y, scoring='neg_mean_squared_error', n_jobs=-1)
    pprint(result)
    """
