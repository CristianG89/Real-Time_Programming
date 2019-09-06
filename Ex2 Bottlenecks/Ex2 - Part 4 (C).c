//https://www.onlinegdb.com/online_c_compiler

#include <pthread.h>	//Threads library
#include <stdio.h>

pthread_t thread_1;		//References to 2nd & 3st thread
pthread_t thread_2;

pthread_mutex_t lock;	//Reference to the mutex

int i = 0;				//Global variable

void *inc() {
    pthread_mutex_lock(&lock);	//Access to var i is now locked (other threads must wait to modify var i)
	for (int j=0; j<1000000; j++) {
		i += 1;
	}
	pthread_mutex_unlock(&lock);	//Access to var i is now unlocked (other threads can start to modify var i)
}

void *dec() {
    pthread_mutex_lock(&lock);
	for (int j=0; j<1000000; j++) {
		i -= 1;
	}
	pthread_mutex_unlock(&lock);
}

int main() {
    pthread_mutex_init(&lock, NULL);	//Mutex is initialized

	pthread_create(&thread_1, NULL, inc, NULL);	//2nd & 3st thread are executed, executing
	pthread_create(&thread_2, NULL, dec, NULL);	//functions inc/dec and passing no argument (NULL)
    
	pthread_join(thread_1, NULL);	//Wait for 2nd thread to finish
	pthread_join(thread_2, NULL);	//Wait for 3st thread to finish
	
	pthread_mutex_destroy(&lock);	//Mutex is destroyed
	
	printf("Final value: %d\n", i);	//The final value is printed

	return 0;
}