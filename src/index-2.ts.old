import inquirer, { Answers } from "inquirer";
import * as fs from "fs";
import simpleGit, { SimpleGit } from 'simple-git';

interface LocationOptions {
    options: string[];
    layer: number;
}

let git: SimpleGit | null = null;

async function recurse(location: LocationOptions): Promise<void> {
    try {
        const answers: Answers = await inquirer.prompt<Answers>([
            {
                type: 'list',
                name: 'entrypoint',
                message: 'What do you want to do?',
                choices: location.options,
            },
        ]);

        if (answers.entrypoint === "init") {
            if (!fs.existsSync("/home/spiderunderurbed/.config/nixos-git-deploy")) {
                fs.mkdirSync("/home/spiderunderurbed/.config/nixos-git-deploy");
                git = simpleGit('/home/spiderunderurbed/.config/nixos-git-deploy');
                await git.init();
                await git.addRemote('origin', '');
            }
        } else if (answers.entrypoint === "apply") {
            // Add your logic for "apply" here
        } else if (answers.entrypoint === "remove") {
            // Add your logic for "remove" here
        } else if (answers.entrypoint === "upgrade") {
            // Add your logic for "upgrade" here
        } else if (answers.entrypoint === "status") {
            // Add your logic for "status" here
        }
    } catch (error) {
        handleGenericError(error);
    }
}

function handleGenericError(error: Error): void {
    console.error('Generic error:', error.message);
}

recurse({ options: ["init", "apply", "status", "remove", "upgrade"], layer: 1 });
