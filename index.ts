import * as functions from 'firebase-functions'
import { WebhookClient } from 'dialogflow-fulfillment'

import { google, calendar_v3 } from 'googleapis'

process.env.DEBUG = 'dialogflow:debug' // enables lib debugging statements

const serviceAccount = {
  type: 'service_account',
  project_id: 'kkweon-free-tier',
  private_key_id: 'b2015d975ac8dfa4e5162f1e4ee891f4cbf7545d',
  private_key:
    '-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCRpjRzHGFAuif4\nKFgmywTDyXxbmLP5eTncTQPgMIuAaORw2plQRYWi8fmC6Jq+TEwY3bilbK1jejxf\nUJ8I0uxmilbFQ+jsTkx4M/glJzYtAWgHsCyCTSPXn2i7KEjZphkl6HbR2Ur67ut/\niIWMSlISaqk1wLkWmA9xa0ZUqZ0oAdpuM0gjZNF4eMIGLlLTecyQ1H7LU81HRh0s\nMU3Cq5zC9fra1AURcagzzpCoFUtKZhtrNtLzhPiQesnwfjtxTRlTLL446r+9nxdg\nuqDnvqXIFZ9APWtWUINu6z0O83WlwBDgD/fmwVGNVPFc0Jb/swGcB4i79WodTPVn\nijpx8rHLAgMBAAECggEADIbLZwUPVjZMfrrVjgyS7dXT5LrW68Nh1xEmnq7+KH6c\n+xnJ6s335uJBz+D6ghhkyRS0r0GQDgiyzY3NB8DARTdrrA6hp0U6rXHmyyc6sRc4\nX8TmpxREW2Sh6MrXrSRscEa0hWrXWIqY5YCT39N6iIv03qMjKA7O8TXFOD3YPtr8\nt0T58CvNLchpgZBPjMiVxT7l0icym+TJQx4DYOYvDoJftxBP/RpcpCrGb4xbqxrD\na+3y32SpnKD3z7NO1LRjmctUgZeQi3WUSXwZaHO7QzMSJ+M5XFUifqPHaL1ZbBuM\n/Sr5WpqSKRVi3mbadHmkYDrj2XTDBQB9LwJ8qYpQjQKBgQDHf9uqEOuJCp15Jrgn\nifjbl4BdC99CpzLxtUD4fZAkCG2fRJaubCSapVeQftvkWbOidqSI/HG1HhwihnR6\nBVvlXv0NIAQ75Ux2YFdcdTpp5qrh/OT2pSCNufeERkgsDa+JPrYro2qqnj8QDqVO\nquP5LfImgxu/RbNKbuz4uLdr3wKBgQC65hcRljijCOM+EH+5dvYSkYR/zoiDfY/Q\nJ9SVLWs6YiERA09vbNpWitXOe0JLU977LvZmVgVAb5h5YApppnnxhXWwvMn8c70I\nXtzm/zdJH4LivCnePdkPT24H7KDsvQN3uVVSvjLCeRwDg9IWbKYeKx/8ODtolnKI\nFwlpmXw3lQKBgQCY1XPUWrgGubIgITCNYd7bY7o8Dh9Q8cROdbw7Yf4uDKLmk+YX\n49M7AhYOJZGR48KBYQD1zOfTiCinrnfHDxnyo42bI37639RvD6l9tHU2sjcRf+ts\npN5GlURw+mLKFQX4T6nBzqSl5yuKwp2TocmamL9dD64PH3eWO1qhxOkH4wKBgDdL\n6Dtd1Lf34zPzsbZvyfJId4lQ0/cDaU9O2YihfX4yllHwRspSzG6aeRO0SDL9R5XN\nmT1B6h/cZKJUlgAYLzAUKnP2B1TX8W/OkVEO5Y6O8iyfO0vzxIrRF17k1d/1NFdx\n0BrBB0eeiXlIwRm9X5DBdZ8sC/evu4ckObayoJvZAoGAatnfH3Iyg+C3jJ3k5mos\nxkjrWZb8cJbDPcGrrBPPgAvb4PgjTxBK2Ls8zrFm7upCUN4+b4tpz3F/Jy29pviu\nt32ZgPd674ezgllGd6sOCdkEVALA8SRzdFk4TVq2JHJdM6IOHxAMcAc/E/TMeyGW\n2Gk95/e+05Ydkw9poVvL1ls=\n-----END PRIVATE KEY-----\n',
  client_email:
    'family-calendar-scheduler@kkweon-free-tier.iam.gserviceaccount.com',
  client_id: '102892603411663619603',
  auth_uri: 'https://accounts.google.com/o/oauth2/auth',
  token_uri: 'https://oauth2.googleapis.com/token',
  auth_provider_x509_cert_url: 'https://www.googleapis.com/oauth2/v1/certs',
  client_x509_cert_url:
    'https://www.googleapis.com/robot/v1/metadata/x509/family-calendar-scheduler%40kkweon-free-tier.iam.gserviceaccount.com',
}
const calendarId = 'l9c4qhf35d3102u4mmsmlggeqo@group.calendar.google.com'

const serviceAccountAuth = new google.auth.JWT({
  email: serviceAccount.client_email,
  key: serviceAccount.private_key,
  scopes: 'https://www.googleapis.com/auth/calendar',
})

const calendar: calendar_v3.Calendar = google.calendar({
  version: 'v3',
  auth: serviceAccountAuth,
})

export const dialogflowFirebaseFulfillment = functions.https.onRequest(
  (request, response) => {
    const agent = new WebhookClient({ request, response })
    console.log(
      'Dialogflow Request headers: ' + JSON.stringify(request.headers),
    )
    console.log('Dialogflow Request body: ' + JSON.stringify(request.body))
    console.log('Parameters' + JSON.stringify(agent.parameters))

    function welcome(agent: WebhookClient) {
      agent.add('Welcome to my agent!')
    }

    function fallback(agent: WebhookClient) {
      agent.add("I didn't understand")
      agent.add("I'm sorry, can you try again?")
    }

    // eslint-disable-next-line
    function handleCreateCalendar(agent: WebhookClient) {}

    const intentMap = new Map()
    intentMap.set('Default Welcome Intent', welcome)
    intentMap.set('Default Fallback Intent', fallback)
    intentMap.set('Create a new event', handleCreateCalendar)
    agent.handleRequest(intentMap)
  },
)

// eslint-disable-next-line @typescript-eslint/no-unused-vars
function createCalendarEvent(
  dateTimeStart: Date,
  dateTimeEnd: Date,
  title: string,
  description: string,
) {
  return new Promise((resolve, reject) => {
    calendar.events.list(
      {
        calendarId,
        timeMin: dateTimeStart.toISOString(),
        timeMax: dateTimeEnd.toISOString(),
      },
      (err, calendarResponse) => {
        // Check if there is a event already on the Calendar
        if (
          err ||
          (calendarResponse != null &&
            calendarResponse.data.items != null &&
            calendarResponse.data.items.length > 0)
        ) {
          reject(
            err ||
              new Error('Requested time conflicts with another appointment'),
          )
        } else {
          // Create event for the requested time period
          calendar.events.insert(
            {
              calendarId: calendarId,
              requestBody: {
                guestsCanModify: true,
                summary: title,
                description: description,
                start: {
                  dateTime: dateTimeStart.toISOString(),
                },
                end: {
                  dateTime: dateTimeEnd.toISOString(),
                },
              },
            },
            (err, event) => {
              err ? reject(err) : resolve(event)
            },
          )
        }
      },
    )
  })
}
