package client

import (
	"fmt"
	"log"
	"time"
)

func ExampleDial() {
	conn, err := Dial("localhost:9851")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	resp, err := conn.Do("set fleet truck1 point 33.5123 -112.2693")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(resp))
}

func ExampleDialPool() {
	pool, err := DialPool("localhost:9851")
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	// We'll set a point in a background routine
	go func() {
		conn, err := pool.Get() // get a conn from the pool
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close() // return the conn to the pool
		_, err = conn.Do("set fleet truck1 point 33.5123 -112.2693")
		if err != nil {
			log.Fatal(err)
		}
	}()
	time.Sleep(time.Second / 2) // wait a moment

	// Retrieve the point we just set.
	go func() {
		conn, err := pool.Get() // get a conn from the pool
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close() // return the conn to the pool
		resp, err := conn.Do("get fleet truck1 point")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(resp))
	}()
	time.Sleep(time.Second / 2) // wait a moment
}
