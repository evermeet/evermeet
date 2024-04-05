
export function makeCollections (api) {
    return {
        events: {
            schema: {
                version: 0,
                primaryKey: 'id',
                type: 'object',
                properties: {
                    id: {
                        type: 'string',
                        maxLength: 32,
                    },
                    name: {
                        type: 'string'
                    }
                },
                encrypted: [
                    'name'
                ]
            },
            methods: {
                async view () {
                    const json = this.toJSON()
                    if (json.calendarId) {
                        const calendar = await api.cols.calendars.findOne({ selector: { id: json.calendarId }}).exec()
                        json.calendar = await calendar.view()
                    }
                    return json;
                },
                async view_audit () {
                    return []
                }
            },
        },
        calendars: {
            schema: {
                version: 0,
                primaryKey: 'id',
                type: 'object',
                properties: {
                    id: {
                        type: 'string',
                        maxLength: 32
                    },
                    name: {
                        type: 'string'
                    }
                }
            },
            methods: {
                async view () {
                    const json = this.toJSON()
                    return json;
                }
            },
        },
        users: {
            schema: {
                version: 0,
                primaryKey: 'id',
                type: 'object',
                properties: {
                    id: {
                        type: 'string',
                        maxLength: 32
                    },
                    name: {
                        type: 'string'
                    },
                    emails: {
                        type: 'array',
                        items: {
                            type: 'string',
                            format: 'email'
                        }
                    },
                    emailShas: {
                        type: 'array',
                        items: {
                            type: 'string'
                        }
                    }
                },
                encrypted: [
                    'emails'
                ]
            }
        },
        sessions: {
            schema: {
                version: 0,
                primaryKey: 'id',
                type: 'object',
                properties: {
                    id: {
                        type: 'string',
                        maxLength: 42
                    },
                    user: {
                        type: 'string'
                    },
                    expiry: {
                        type: 'string'
                    },
                    time: {
                        type: 'string'
                    }
                }
            }
        }
    }
}