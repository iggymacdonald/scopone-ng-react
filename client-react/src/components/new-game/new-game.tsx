import { Card, CardHeader, CardContent, TextField } from "@material-ui/core";
import { ChangeEvent, useContext, useState } from "react";
import { ServerContext } from "../../context/server-context";
import { useStyles } from "../style-material-ui";

import "../style.css";

export const NewGame = () => {
  const [gameName, setGameName] = useState("");
  const server = useContext(ServerContext);

  const classes = useStyles();

  const handleChange = (
    event: ChangeEvent<HTMLTextAreaElement | HTMLInputElement>
  ) => {
    setGameName(event.target.value);
  };

  const keyPress = (e: { key: string }) => {
    if (e.key === "Enter") {
      server.newGame(gameName);
    }
  };

  return (
    <Card>
      {/* https://material-ui.com/customization/components/#overriding-styles-with-global-class-names */}
      {/* <CardHeader title="Player" classes={{ title: "header" }}></CardHeader> */}
      {/* https://material-ui.com/customization/components/#overriding-styles-with-class-names */}
      <CardHeader title="Create New Game" className="header"></CardHeader>
      <CardContent>
        <div className={classes.root}>
          <TextField
            id="standard-full-width"
            label="Name"
            style={{ margin: 8 }}
            placeholder="Your name as player"
            helperText="Use the same name if you loose session and want to riconnect"
            fullWidth
            margin="normal"
            InputLabelProps={{
              shrink: true,
            }}
            onChange={handleChange}
            onKeyPress={keyPress}
          />
        </div>
      </CardContent>
    </Card>
  );
};
