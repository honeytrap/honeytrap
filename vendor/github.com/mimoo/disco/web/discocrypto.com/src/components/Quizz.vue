<template>
  <article class="message">
  <div class="message-body">


    <div class="field displayBox">
      <div v-if="finished">


          <div class="box" v-for="result in results">
              <div class="content">
                <p>
                  <strong><router-link :to="'/protocol/' + result.name">{{result.name}}</router-link></strong> <span class="tag" v-for="tag in result.tags">{{tag}}</span>
                  <div v-html="result.description"></div>
                </p>
              </div>
          </div>

      </div>
      <strong v-else><i class="fa fa-question-circle" aria-hidden="true"></i> {{question}}</strong>
    </div>


    <div class="field is-grouped">

      <div class="control">
    <button href="some/link" class="button is-success"  v-on:click="quizzYes" :disabled="finished">
    <span class="icon is-small">
        <i class="fa fa-check"></i>
    </span>
</button>
</div>
      <div class="control">

    <button href="some/link" class="button is-danger"  v-on:click="quizzNo" :disabled="finished">
    <span class="icon is-small">
        <i class="fa fa-times"></i>
    </span>
</button>
</div>
      <div class="control">
        <button class="button is-text" v-on:click="resetQuizz">Reset the Quizz</button>
      </div>
    </div>

    <hr>

    <progress class="progress is-small" :value="questionId" max="100"></progress>

  </div>
  </article>
</template>

<style scoped>
.displayBox{
  margin-bottom:20px;
}
progress {
  transition: all 0.5s ease;
}
</style>

<script>
  import patterns from '@/assets/patterns.json';

  // this quizz works mostly by elimination.
	var questionsreponses = [
		{
			q: "Does your protocol involve only clients talking to a server (where the server doesn't reply back)?",
			ryes: "one_way",
      rno: "two_way",
		},
    {
    	q: "Does your protocol involve a client and a server already sharing a secret (a symmetric key)?",
    	ryes: "psk"
    },
    {
   		q: "Does at least one of your peer already know the public key of the other peer?",
   		ryes: "key_pinning"
   	},
    {
   		q: "Is at least one of your peer's public key signed by a trusted authority?",
   		ryes: "PKI"
   	},
    {
   		q: "Do your peers have an out-of-band way to compare a string or a large number?",
   		ryes: "out_of_band"
   	},
  ]

  export default {
    name: 'Quizz',
    data () {
      return {
        question: 'wait',
        questionId: 0,
        tags: [],
        finished: false,
        results: [],
      }
    },

    methods: {
      resetQuizz (e) {
        this.results = []
        this.questionId = 0
        this.question = questionsreponses[0]["q"]
        this.tags = []
        this.finished = false
        e.preventDefault()
      },
      quizzNext() {
        this.questionId++
        if(this.questionId >= questionsreponses.length) {
          this.displayResults()
          return
        }
        this.question = questionsreponses[this.questionId]["q"]
      },
      quizzYes (e) {
        if(questionsreponses[this.questionId].hasOwnProperty("ryes")){
          this.tags.push(questionsreponses[this.questionId]["ryes"])
        }
        this.quizzNext()
        e.preventDefault()
      },
      quizzNo (e) {
        if(questionsreponses[this.questionId].hasOwnProperty("rno")){
          this.tags.push(questionsreponses[this.questionId]["rno"])
        }
        this.quizzNext()
        e.preventDefault()
      },
      displayResults() {
        this.finished = true
        this.question = this.tags
        patterns.forEach( (pattern) => {
          var fitting = true
          pattern.tags.forEach( (tag) => { 
            console.log("test", tag, "for", this.tags, " result:", this.tags.includes(tag))
            if( !this.tags.includes(tag) ) { fitting = false }
          })
          if(fitting) {
            console.log("fit:", pattern)
            this.results.push(pattern)
          }
        })
        if(this.results.length == 0) {
          this.results.push({name: "No pattern matching", description: "You need to answer yes to some questions :)"})
        }
      }
    },

    mounted () {
        this.question = questionsreponses[0]["q"]
        document.querySelector("progress").max = questionsreponses.length
    }
  }
</script>