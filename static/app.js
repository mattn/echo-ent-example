const { createApp } = Vue;

createApp({
  data() {
    return {
      comments: [],
      name: "",
      text: "",
    };
  },
  methods: {
    add: function () {
      const payload = { name: this.name, text: this.text };
      axios
        .post("/api/comments", payload)
        .then(() => {
          this.name = "";
          this.text = "";
          this.update();
        })
        .catch((err) => {
          alert(err.response.data.message);
        });
    },
    update: function () {
      axios
        .get("/api/comments")
        .then((response) => (this.comments = response.data || []))
        .catch((error) => console.log(error));
    },
  },
  created() {
    this.update();
  },
}).mount("#app");
