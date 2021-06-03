# Stolen_Cars_Reporting_System


Overview and Approach to the problem :

ReportStolenCars is a website where users can provide details about their stolen cars. Police officers are assigned to each case which is registered on the website. Users can track the status of their cases on the site. New cases are automatically assigned to an unoccupied officer, and 1 officer can attend to only 1 case.

Key approaches to mapping the cases to the officers:

1) If a case is registered on the site and an officer is free, the case is assigned to the officer and he cannot take on any new cases until the current one is resolved.
2) If there are x cases and x occupied officers the x+1th case is tagged unassigned until a case gets resolved or an officer is added to the force, in which case the x+1th case is assigned to the officer.
3) If there are multiple unassigned cases and 1 unoccupied officer, the earliest registered case is assigned to the officer.
4) If there are multiple unoccupied officers and 1 unassigned case, the case is assigned to the officer which was least recently occupied.

Technologies used: 
1) Flask, JavaScript (front end)
2) Golang, Gorilla mux (framework like flask) backend.
3) MongoDB database.

Steps to Run the application :
Install MongoDB 
Download python 3.5.x or above for your machine.
Unzip the file in the directory of your choice
Run all the following commands from the root of the unzipped directory.
Run ‘pip install -r requirements.txt’.
Run ‘python api.py’ which runs on localhost:3000
Run the ‘app.exe’  binary, if the machine is not windows or you wish to build the binary in your system, follow the steps below.
Install Golang 1.12.x+ on your system.
If your machine is windows, set all the path variables correctly (GOPATH and GOROOT).
Run ‘go get github.com/gorilla/mux’
Run ‘go get go.mongodb.org/mongo-driver’ , ignore the go warning
Run ‘go build -o <file-name>’
Run ‘start <file-name>.exe’ for windows and ‘./<file-name>’ for Linux. The application runs on port 8080.
Type localhost:3000 in your browser URL and the website is ready for exploration
