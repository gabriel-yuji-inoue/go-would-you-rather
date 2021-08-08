scope = {
  question: {},
  initialize: function() {
    const self = scope
    self.loadQuestion()
  },
  loadQuestion: function(){
    const self = scope

    self.template.enableContainerOptions(false)
    $.get("/question", async function(data) {
      self.question = data
      self.template.renderOptions(self.question)
    })
  },
  choose: function(option) {
    const self = scope
    scope.template.enableOptions(false)

    $.post("/question/answer", {
      "id": self.question.id,
      "option": option
    }, function( data ) {
      self.question = data
      scope.template.renderOptionsResult(self.question, option)
    });
    
    
  },
  template: {
    enableOptions: (enable = true) => {
      if (enable) {
        $("#first-option").removeClass("disabled")
        $("#second-option").removeClass("disabled")
      } else {
        $("#first-option").addClass("disabled")
        $("#first-option").html(`<i class="fas fa-sync fa-spin"></i>`)
        $("#second-option").addClass("disabled")
        $("#second-option").html(`<i class="fas fa-sync fa-spin"></i>`)
      }
    },
    enableContainerOptions: (enable = true) => {
      const self = scope.template
      if (enable) {
        $("#card-body").removeClass("opacity-20")
        $("#i-load-question").removeClass("d-none")
        $("#i-loading").addClass("d-none")
        $("#load-question-button").removeClass("disabled")
        self.enableOptions()
      } else {
        $("#card-body").addClass("opacity-20")
        $("#i-load-question").addClass("d-none")
        $("#i-loading").removeClass("d-none")
        $("#load-question-button").addClass("disabled")
        self.enableOptions(false)
      }
    },
    renderOptions: (question) => {
      const self = scope.template

      self.enableOptions()
      self.enableContainerOptions()
      $("#title").html(question.title)
      $("#first-option").html(question.first_option_description)
      $("#second-option").html(question.second_option_description)
      $("#details").html(question.details)
    },
    renderOptionsResult: (questionResult, choosed) => {      
      loadPercent = async (per, q) => {
        for (i = 0; i <= per; i++) {
          await scope.utils.sleep(25)
          $(q).html(`${i}%`)
        }
      }
      allVotes = questionResult.first_option_votes + questionResult.second_option_votes
      first_option_votesPercent = Math.ceil(questionResult.first_option_votes/allVotes * 100)
      second_option_votesPercent = 100 - first_option_votesPercent

      templateFirstOption = `
      <div class="text-center">
        <span class="fs-4"><span id="first-option-percent">-%</span> ${choosed=='first-option' ? "agree" : "disagree"}</span><br />
        <span>${questionResult.first_option_votes}</span><br />
        <span>${questionResult.first_option_description}</span>
      </div>
      `
      templateSecondOption = `
      <div class="text-center">
        <span class="fs-4"><span id="second-option-percent">-%</span> ${choosed=='second-option' ? "agree" : "disagree"}</span><br />
        <span>${questionResult.second_option_votes}</span><br />
        <span>${questionResult.second_option_description}</span>
      </div>
      `
      
      $("#first-option").html(templateFirstOption)
      loadPercent(first_option_votesPercent, "#first-option-percent")
      $("#second-option").html(templateSecondOption)
      loadPercent(second_option_votesPercent, "#second-option-percent")
    }
  },
  utils: {
    sleep: function(ms) {
      return new Promise(resolve => setTimeout(resolve, ms));
    }
  }
}


setTimeout(() => {
  scope.initialize()
})