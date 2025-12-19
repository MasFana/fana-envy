import time
import os

APP_ENV = os.environ.get("APP_ENV")
DB_HOST = os.environ.get("DB_HOST")
DEBUG = os.environ.get("DEBUG")

print("ENVs:", { APP_ENV, DB_HOST, DEBUG });

while True:
    print("Hello Python")
    time.sleep(3)

