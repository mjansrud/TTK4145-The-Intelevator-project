By Morten Jansrud and Endre Gjølstad!
Uuntz.

Explanation of packages:
-----------------------

Main: The executing package upon program start.  
Startup: Guarantees a successfull startup, creates all necessary threads and a backup terminal  
Driver: Is a set of functions to control the elevator  
Hardware: Lowest level interface to interact with the elvator  
Events: Event loops, the brain of the system  
Orders: Handles everything regarding the orders.  
Network: The network module allowing communication between elevators.  
States: Tracks the local elevator states and manages the connected elevators.  
Utilities: Adds the structs and definitions used throughout the program.  
