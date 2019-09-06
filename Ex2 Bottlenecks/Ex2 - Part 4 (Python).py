#https://www.onlinegdb.com/online_python_compiler

import threading
from threading import Thread	#Threads library
from threading import Lock		#Locks (mutex) library

lock = Lock() #An object of type Lock is created to handle mutual exclusion (mutex)

#In Python you "import" a global variable, instead of "export"ing it when you declare it.
#This is the reason why i is declared as global inside each function where it will be used.
i = 0				#Global variable

def inc():		#Function "inc"
	global i	#This function uses variable i, which has been declared and initialized globally
	lock.acquire()	#Access to var i is now locked (other threads must wait to modify var i)
	
	for x in range(1000000):
	    i += 1
	
	lock.release()	#Access to var i is now unlocked (other threads can start to modify var i)

def dec():		#Function "dec"
	global i
	lock.acquire()
	
	for x in range(1000000):
	    i -= 1
	
	lock.release()
    
def main():
    thread_1 = threading.Thread(target=inc, args = ())	#New thread (2nd) will execute function "inc"
    thread_2 = threading.Thread(target=dec, args = ())	#New thread (3rd) will execute function "dec"
    
    thread_1.start()	#Threads are executed
    thread_2.start()
    
    thread_1.join()		#The main thread waits until threads "thread_1" and "thread_2" finish
    thread_2.join()
    
    print("The final number is %d" % (i))	#Final value of i is printed

main()		#Execute function main