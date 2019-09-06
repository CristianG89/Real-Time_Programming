#https://www.onlinegdb.com/online_python_compiler

import threading	#Threads library

#In Python you "import" a global variable, instead of "export"ing it when you declare it.
#This is the reason why i is declared as global inside each function where it will be used.
i = 0				#Global variable

def inc():		#Function "inc"
	global i	#This function uses variable i, which has been declared and initialized globally
	for x in range(1000000):
		i += 1

def dec():		#Function "dec"
	global i
	for x in range(1000000):
		i -= 1

thread_1 = threading.Thread(target=inc, args = ())	#New thread (2nd) will execute function "inc"
thread_2 = threading.Thread(target=dec, args = ())	#New thread (3rd) will execute function "dec"

thread_1.start()	#Threads are executed
thread_2.start()

thread_1.join()		#The main thread waits until threads "thread_1" and "thread_2" finish
thread_2.join()

print("The final number is %d" % (i))	#Final value of i is printed