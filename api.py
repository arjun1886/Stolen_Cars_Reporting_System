from flask import Flask,Blueprint, render_template, request, jsonify, redirect, url_for
from flask_login import login_user,logout_user,login_required
import hashlib
import base64
import binascii
import json
import datetime
from pathlib import Path
import requests
from flask_cors import CORS
import os

app = Flask(__name__)
CORS(app)


@app.route('/')
def index():
  return render_template('index.html')

@app.route('/report')
def report():
    return render_template('report.html')  

@app.route('/track')
def track():
    return render_template('track.html')  

if __name__ == '__main__':
   app.run(host='localhost',port="3000")

    

    



  

  

