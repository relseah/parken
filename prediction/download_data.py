import os
import requests
from html.parser import HTMLParser

from prediction import HISTORICAL_DATA_DIRECTORY


class DataIDHTMLParser(HTMLParser):
    def __init__(self):
        self.current_id = None
        self.parse_parking_name = False
        self.data_ids = {}
        super().__init__()

    def handle_starttag(self, tag, attrs):
        for name, value in attrs:
            if name == 'data-id':
                self.current_id = value
            elif self.current_id and name == 'class' and value == 'description':
                self.parse_parking_name = True

    def handle_data(self, data):
        if self.parse_parking_name:
            self.data_ids[data.strip()] = self.current_id
            self.current_id = None
            self.parse_parking_name = False


r = requests.get(
    'https://ckan.datenplattform.heidelberg.de/de/dataset/mobility_main_parking')
parser = DataIDHTMLParser()
parser.feed(r.text)

folder = 'historical_data'
if not os.path.exists(folder):
    os.mkdir(folder)
for name, data_id in parser.data_ids.items():
    # dirty
    # not tested, how os.path.join would handle slashes in name
    path = os.path.join(HISTORICAL_DATA_DIRECTORY,
                        name.replace('/', ' ') + '.csv')
    if os.path.exists(path):
        continue
    r = requests.get(
        'https://ckan.datenplattform.heidelberg.de/de/datastore/dump/{}?bom=True'.format(data_id))
    with open(path, "w", encoding='utf-8') as file:
        file.write(r.text)
