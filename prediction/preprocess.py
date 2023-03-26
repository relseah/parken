import os
import csv
import re
import sys
from datetime import datetime

from prediction import HISTORICAL_DATA_DIRECTORY, generate_sample


class Occupancy:
    def __init__(self, time, spots):
        self.time, self.spots = time, spots

    def sample(self, parking_id):
        return generate_sample(parking_id, self.time)


def save(filename, rows):
    with open(filename, 'w', newline='') as file:
        writer = csv.writer(file)
        writer.writerows(rows)


if __name__ == '__main__':
    pattern = re.compile('P(\d+) -')
    raw_data = {}
    for filename in os.listdir(HISTORICAL_DATA_DIRECTORY):
        match = pattern.match(filename)
        if match is None:
            sys.exit("invalid filename for dump: " + filename)
        with open(os.path.join(HISTORICAL_DATA_DIRECTORY, filename), encoding='utf-8') as file:
            occupancies = [Occupancy(datetime.fromisoformat(raw_occupancy['recvTime']), int(raw_occupancy['availableSpotNumber']))
                           for raw_occupancy in csv.DictReader(file)]
            raw_data[int(match.group(1))] = occupancies
    data = []
    target = []
    for parking_id, occupancies in raw_data.items():
        for occupancy in occupancies:
            data.append(occupancy.sample(parking_id))
            target.append(occupancy.spots)
    save('data.csv', data)
    save('target.csv', [target])
