from tkinter.tix import Select
from xmlrpc.client import DateTime
import requests, random

# Base URL of the service
base_url = "http://127.0.0.1:8090"

# Function to make a POST request to /authentication/login/{username}/{password}
def test_login(username, password):
    url = base_url + f"/authentication/login/{username}/{password}"
    response = requests.post(url)
    print(f"* AUTH: {response}")
    #println("POST", url, response)
    return response

# Function to make a POST request to /authentication/logout
def test_logout():
    url = base_url + "/authentication/logout"
    response = requests.post(url)
    println("POST", url, response)

# Function to make a GET request to /meters/setting-result/{acceptNo}
def test_load_survey_result(acceptNo):
    url = base_url + f"/meters/setting-result/{acceptNo}"
    response = requests.get(url)
    println("GET", url, response)

# Function to make a GET request to /articles/{category}/{id}
def test_get_article(category, article_id):
    url = base_url + f"/articles/{category}/{article_id}"
    response = requests.get(url)
    println("GET", url, response)
    
# Function to print request information in a formatted way
def aprintln(method, url, response):
    print(f"{'='*50}")
    print(f"{method} {url}")
    if response:
        print(f"Status Code: {response.status_code}")
        try:
            json_response = response.json()
            print("Response JSON:")
            print(json_response)
        except ValueError:
            print("Response Content:")
            print(response.content.decode('utf-8'))
    else:
        print("No response received.")

def println(a,b,abcd):
    abc = abcd.json()
    defi = str(abc["_responce"])
    print(f"<return> {defi}")
########################
### SOCIAL MEDIA APP ###
########################

# Function to make a POST request to /authentication/login/{username}/{password}
def PostArticle(ssid,post):
    url = base_url + f"/post/article/{ssid}/{post}"
    response = requests.post(url)
    println("POST", url, response)


# Testing the endpointss
import time, datetime, os
if __name__ == "__main__":
    st = datetime.datetime.now()
    tot_req = 10
    # for _ in range(tot_req):
    #     print(_,"\n")
    #     test_login("hamza", "hamza")
    #     # print()
    #     # test_logout()
    #     #time.sleep(3)
    resp = test_login("hamza","hamza")
    respo = resp.json()
    ssid = respo["sessionId"]
    PNums = [10,12,19,22,25,29,32,34,36,39,44,46,50,55,69,79,125]
    #SN = random.choice(PNums)
    #print(SN)
    for _ in range(200):
        PostsList = ["Hello, good to see you all!","Hey who are you?","Yo Yo Yo, whats up dude?","Hi i am hamza and i am almost 16 years old."]
        SelectedPost = input(">>> ") #random.choice(PostsList)
        PostArticle(ssid,f"{SelectedPost}")
        time.sleep(1)
    et = datetime.datetime.now()
    print(f"{tot_req} Requests sended: Time taken: {et-st}, with 1sec pause after each post.")
    
    

# print()
# test_load_survey_result("123")
# print()
# test_get_article("tech", "456")
